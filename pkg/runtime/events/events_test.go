package events

import (
	"testing"
	"time"
)

func TestPublishAndSubscribe(t *testing.T) {
	Reset()
	ch, cancel := Subscribe(4)
	defer cancel()

	evt := ServiceAction{TS: time.Now(), ServiceID: "web", Message: "started"}
	Publish(evt)

	select {
	case received := <-ch:
		if act, ok := received.(ServiceAction); !ok {
			t.Fatalf("expected ServiceAction, got %T", received)
		} else if act.Message != evt.Message || act.ServiceID != evt.ServiceID {
			t.Fatalf("unexpected payload: %#v", act)
		}
	case <-time.After(time.Second):
		t.Fatalf("expected event delivery")
	}
}

func TestPublishDropOnFullBuffer(t *testing.T) {
	Reset()
	ch, cancel := Subscribe(1)
	defer cancel()

	Publish(ServiceAction{TS: time.Now(), ServiceID: "a", Message: "1"})
	Publish(ServiceAction{TS: time.Now(), ServiceID: "a", Message: "2"})

	<-ch
	select {
	case <-ch:
		t.Fatalf("buffer should have dropped second event")
	default:
	}
}
