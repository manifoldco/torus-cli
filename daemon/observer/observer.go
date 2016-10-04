// Package observer provides a facility for publishing progress updates and
// state changes from parts of the daemon, an a SSE http handler for consumers
// of these events.
package observer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

type ctxkey string

// CtxRequestID is the context WithValue key for a request id.
var CtxRequestID ctxkey = "id"

// EventType represents all possible types of events that can be observed.
type EventType string

// All values for EventType
const (
	Progress EventType = "progress"
	Started  EventType = "started"
	Finished EventType = "finished"
	Errored  EventType = "errored"
	Aborted  EventType = "aborted"
)

type event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"-"` // type is included in the SSE
	Message   string    `json:"message"`
	Completed uint      `json:"completed"`
	Total     uint      `json:"total"`
}

type notification struct {
	Type      EventType
	Message   string
	Increment bool
}

// Observer recieves events via Notify, and publishes them as SSEs via its
// ServeHTTP function.
type Observer struct {
	notify chan *event
	closed chan int

	observers map[chan []byte]bool

	newObservers    chan chan []byte
	closedObservers chan chan []byte
}

type transaction struct {
	requestID      string
	total          uint
	current        uint
	events         chan<- *event
	totalUpdates   <-chan uint
	notifications  <-chan *notification
	observerClosed <-chan int
	ctxDone        <-chan struct{}
}

// Notifier belongs to a transactions and represents one segment in a series of
// actions. A Notifier can send many messages.
type Notifier struct {
	total          uint
	current        uint
	transaction    *transaction
	totalUpdates   chan<- uint
	notifications  chan<- *notification
	observerClosed <-chan int
	ctxDone        <-chan struct{}
}

func (t *transaction) start() {
	go func() {
		for {
			select {
			case notification := <-t.notifications:
				if notification.Increment {
					t.current++
				}

				evt := &event{
					ID:        t.requestID,
					Type:      notification.Type,
					Message:   notification.Message,
					Completed: t.current,
					Total:     t.total,
				}

				t.events <- evt
			case size := <-t.totalUpdates:
				t.total += size
			case <-t.ctxDone: // transaction has completed
				return
			case <-t.observerClosed: // observer has shutdown
				return
			}
		}
	}()
}

// Notifier creates a child notifier to this Notifier
func (n *Notifier) Notifier(total uint) *Notifier {
	notifier := &Notifier{
		total:          total,
		current:        0,
		transaction:    nil,
		totalUpdates:   n.totalUpdates,
		notifications:  n.notifications,
		observerClosed: n.observerClosed,
		ctxDone:        n.ctxDone,
	}

	n.totalUpdates <- total
	return notifier
}

// Notify publishes an event to all SSE observers.
func (n *Notifier) Notify(eventType EventType, message string, increment bool) {
	notif := &notification{
		Type:      eventType,
		Message:   message,
		Increment: increment,
	}

	if increment {
		n.current++
	}

	if n.current > n.total {
		panic(fmt.Sprintf(
			"notifications exceed maximum %d/%d", n.current, n.total))
	}

	select {
	case n.notifications <- notif:
		return
	case <-n.observerClosed:
		return
	case <-n.ctxDone:
		return
	}
}

// New returns a new initialized Observer
func New() *Observer {
	return &Observer{
		notify: make(chan *event),
		closed: make(chan int),

		observers: make(map[chan []byte]bool),

		newObservers:    make(chan chan []byte),
		closedObservers: make(chan chan []byte),
	}
}

// Notifier creates a new transaction for sending notifications
func (o *Observer) Notifier(ctx context.Context, total uint) (*Notifier, error) {

	if ctx == nil {
		return nil, errors.New("Context must be provided")
	}

	id, ok := ctx.Value(CtxRequestID).(string)
	if !ok {
		return nil, errors.New("Missing 'id' property in Context")
	}

	totalUpdates := make(chan uint)
	notifications := make(chan *notification)

	t := &transaction{
		requestID:      id,
		total:          total,
		current:        0,
		totalUpdates:   totalUpdates,
		events:         o.notify,
		notifications:  notifications,
		observerClosed: o.closed,
		ctxDone:        ctx.Done(),
	}

	t.start()

	n := &Notifier{
		total:          total,
		current:        0,
		transaction:    t,
		totalUpdates:   totalUpdates,
		notifications:  notifications,
		observerClosed: t.observerClosed,
		ctxDone:        t.ctxDone,
	}

	return n, nil
}

// ServeHTTP implements the http.Handler interface for providing server-sent
// events of observed notifications.
func (o *Observer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	rwf := rw.(http.Flusher)
	closed := rw.(http.CloseNotifier).CloseNotify()

	notify := make(chan []byte)
	o.newObservers <- notify

	defer func() {
		o.closedObservers <- notify
	}()

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Connection", "keep-alive")
	rwf.Flush()

	for {
		select {
		case evt := <-notify:
			// Write the event to the client. we ignore errors here, and
			// let the close channel tell us when to stop writing.
			rw.Write(evt)
			rwf.Flush()

		case <-closed: // client has disconnected. Exit.
			return
		case <-o.closed: // The Observer is shutting down. Exit.
			return
		}
	}
}

// Start begins listening for notifications of observable events. It returns
// after stop has been called.
func (o *Observer) Start() {
	for {
		select {
		case evt := <-o.notify: // We have an event to observe
			if len(o.observers) == 0 {
				continue
			}

			evtb, err := json.Marshal(evt)
			if err != nil {
				log.Printf("Error marshaling event: %s", err)
				continue
			}

			sse := []byte("event: " + evt.Type + "\ndata: ")
			sse = append(sse, append(evtb, []byte("\n\n")...)...)

			for n := range o.observers {
				n <- sse
			}

		case n := <-o.newObservers:
			o.observers[n] = true
		case n := <-o.closedObservers:
			delete(o.observers, n)

		case <-o.closed: // The Observer has been closed.
			return
		}
	}
}

// Stop terminates propagation of events through the observer
func (o *Observer) Stop() {
	close(o.closed)
}
