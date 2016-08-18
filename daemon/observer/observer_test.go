package observer

import (
	"bytes"
	"context"
	"net/http/httptest"
	"testing"

	"github.com/satori/go.uuid"
)

func TestObserverNotify(t *testing.T) {
	t.Run("Start does not panic on notify", func(t *testing.T) {
		o := New()
		ctx := context.WithValue(context.Background(), "id", uuid.NewV4().String())

		go o.Start()
		defer o.Stop()

		err := o.Notify(ctx, Progress, "hi", 0, 0)
		if err != nil {
			t.Error("err not nil from Notify")
		}
	})

	t.Run("Notify errors with no id on context", func(t *testing.T) {
		o := New()
		ctx := context.Background() // No id!

		err := o.Notify(ctx, Progress, "hi", 0, 0)
		if err == nil {
			t.Error("err was not set on Notify with no id")
		}
	})
}

type CloseNotifyResponseRecorder struct {
	httptest.ResponseRecorder
	flushed chan bool
}

func (cn *CloseNotifyResponseRecorder) CloseNotify() <-chan bool {
	return make(chan bool)
}

func (cn *CloseNotifyResponseRecorder) Flush() {
	cn.flushed <- true
}

func TestObserverServeHTTP(t *testing.T) {
	t.Run("events are sent via SSE", func(t *testing.T) {
		o := New()
		id := uuid.NewV4().String()
		ctx := context.WithValue(context.Background(), "id", id)

		go o.Start()
		defer o.Stop()

		rw := &CloseNotifyResponseRecorder{
			ResponseRecorder: *httptest.NewRecorder(),
			flushed:          make(chan bool),
		}

		r := httptest.NewRequest("GET", "/observe", nil)

		go o.Notify(ctx, Progress, "hi", 0, 0)
		go o.ServeHTTP(rw, r)

		<-rw.flushed
		<-rw.flushed

		expected := "text/event-stream"
		seen := rw.Header().Get("Content-Type")
		if seen != expected {
			t.Errorf("Unexpected content type. got: %s want: %s", seen, expected)
		}

		expectedEvent := []byte(
			"event: progress\ndata: {\"id\":\"" + id + "\",\"message\":\"hi\",\"completed\":0,\"total\":0}\n\n",
		)
		if !bytes.Equal(rw.Body.Bytes(), expectedEvent) {
			t.Errorf("Event data does not match. got:\n%s\nwanted:\n%s", rw.Body.Bytes(), expectedEvent)
		}
	})
}
