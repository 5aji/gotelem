package gotelem

import (
	"log/slog"
	"math"
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


// makeLiveSystem starts a process that is used to continuously stream
// data into a Broker. Every 100ms it will send either a BmsMeasurement
// or WslVelocity. The values will be static for WslVelocity (to
// make comparison easier) but will be dynamic for BmsMeasurement.
//
func liveStream(done chan bool, broker *Broker) {
	bmsPkt := &skylab.BmsMeasurement{
		Current: 1.23,
		BatteryVoltage: 11111,
		AuxVoltage: 22222,
	}
	wslPkt := &skylab.WslVelocity{
		MotorVelocity: 0,
		VehicleVelocity: 100.0,
	}
	var next skylab.Packet = bmsPkt
	for {
		select {
		case <-done:
			return
		case <-time.After(100 * time.Millisecond):
			// send the next packet.
			if next == bmsPkt {
				bmsPkt.Current = float32(math.Sin(float64(time.Now().Unix()) / 2.0))
				ev := skylab.BusEvent{
					Timestamp: time.Now(),
					Name: next.String(),
					Data: next,
				}
				broker.Publish("livestream", ev)
				next = wslPkt
			} else {
				// send the wsl
				ev := skylab.BusEvent{
					Timestamp: time.Now(),
					Name: next.String(),
					Data: next,
				}
				broker.Publish("livestream", ev)
				next = bmsPkt
			}



		}
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
