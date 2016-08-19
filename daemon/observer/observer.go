// Package observer provides a facility for publishing progress updates and
// state changes from parts of the daemon, an a SSE http handler for consumers
// of these events.
package observer

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

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

// Observer recieves events via Notify, and publishes them as SSEs via its
// ServeHTTP function.
type Observer struct {
	notify chan *event
	closed chan int

	observers map[chan []byte]bool

	newObservers    chan chan []byte
	closedObservers chan chan []byte
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

// Notify publishes an event to all SSE observers.
func (o *Observer) Notify(ctx context.Context, eventType EventType,
	message string, completed, total uint) error {

	id, ok := ctx.Value("id").(string)
	if !ok || id == "" {
		return errors.New("id not found on context")
	}

	o.notify <- &event{
		ID:      id,
		Type:    eventType,
		Message: message,

		Completed: completed,
		Total:     total,
	}

	return nil
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
