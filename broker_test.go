package gotelem

import (
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/kschamplin/gotelem/skylab"
)

func makeEvent() skylab.BusEvent {
	var pkt skylab.Packet = &skylab.BmsMeasurement{
		BatteryVoltage: 12000,
		AuxVoltage:     24000,
		Current:        1.23,
	}
	return skylab.BusEvent{
		Timestamp: time.Now(),
		Name:      pkt.String(),
		Data:      pkt,
	}
}



func TestBroker(t *testing.T) {
	t.Parallel()

	t.Run("test send", func(t *testing.T) {
		flog := slog.New(slog.NewTextHandler(os.Stderr, nil))
		broker := NewBroker(10, flog)

		sub, err := broker.Subscribe("testSub")
		if err != nil {
			t.Fatalf("error subscribing: %v", err)
		}
		testEvent := makeEvent()

		go func() {
			time.Sleep(time.Millisecond * 1)
			broker.Publish("other", testEvent)
		}()

		var recvEvent skylab.BusEvent
		select {
		case recvEvent = <-sub:
			if !testEvent.Equals(&recvEvent) {
				t.Fatalf("events not equal, want %v got %v", testEvent, recvEvent)
			}
		case <-time.After(1 * time.Second):
			t.Fatalf("timeout waiting for packet")
		}

	})
	t.Run("multiple broadcast", func(t *testing.T) {
		flog := slog.New(slog.NewTextHandler(os.Stderr, nil))
		broker := NewBroker(10, flog)
		testEvent := makeEvent()
		wg := sync.WaitGroup{}

		clientFn := func(name string) {
			sub, err := broker.Subscribe(name)
			if err != nil {
				t.Log(err)
				return
			}
			<-sub
			wg.Done()
		}

		wg.Add(2)
		go clientFn("client1")
		go clientFn("client2")

		// yes this is stupid. otherwise we race.
		time.Sleep(10 * time.Millisecond)

		broker.Publish("sender", testEvent)

		done := make(chan bool)
		go func() {
			wg.Wait()
			done <- true
		}()
		select {
		case <-done:

		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for clients")
		}
	})

	t.Run("name collision", func(t *testing.T) {
		flog := slog.New(slog.NewTextHandler(os.Stderr, nil))
		broker := NewBroker(10, flog)
		_, err := broker.Subscribe("collide")
		if err != nil {
			t.Fatal(err)
		}
		_, err = broker.Subscribe("collide")
		if err == nil {
			t.Fatal("expected error, got nil")
		}

	})

	t.Run("unsubscribe", func(t *testing.T) {
		flog := slog.New(slog.NewTextHandler(os.Stderr, nil))
		broker := NewBroker(10, flog)
		ch, err := broker.Subscribe("test")
		if err != nil {
			t.Fatal(err)
		}
		broker.Unsubscribe("test")
		_, ok := <-ch
		if ok {
			t.Fatal("expected dead channel, but channel returned result")
		}
	})
}
