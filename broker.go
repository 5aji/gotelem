package gotelem

import (
	"errors"
	"sync"

	"log/slog"

	"github.com/kschamplin/gotelem/skylab"
)

// Broker is a Bus Event broadcast system. You can subscribe to events,
// and send events.
type Broker struct {
	subs map[string]chan skylab.BusEvent // contains the channel for each subsciber

	logger  *slog.Logger
	lock    sync.RWMutex
	bufsize int // size of chan buffer in elements.
}

// NewBroker creates a new broker with a given logger.
func NewBroker(bufsize int, logger *slog.Logger) *Broker {
	return &Broker{
		subs:    make(map[string]chan skylab.BusEvent),
		logger:  logger,
		bufsize: bufsize,
	}
}

// Subscribe joins the broker with the given name. The name must be unique.
func (b *Broker) Subscribe(name string) (ch chan skylab.BusEvent, err error) {
	// get rw lock.
	b.lock.Lock()
	defer b.lock.Unlock()
	_, ok := b.subs[name]
	if ok {
		return nil, errors.New("name already in use")
	}
	b.logger.Info("subscribe", "name", name)
	ch = make(chan skylab.BusEvent, b.bufsize)

	b.subs[name] = ch
	return
}


// Unsubscribe removes a subscriber matching the name. It doesn't do anything
// if there's nobody subscribed with that name
func (b *Broker) Unsubscribe(name string) {
	// remove the channel from the map. We don't need to close it.
	b.lock.Lock()
	defer b.lock.Unlock()
	b.logger.Debug("unsubscribe", "name", name)
	if _, ok := b.subs[name]; ok {
		close(b.subs[name])
		delete(b.subs, name)
	}
}

// Publish sends a bus event to all subscribers. It includes a sender
// string which prevents loopback.
func (b *Broker) Publish(sender string, message skylab.BusEvent) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	b.logger.Debug("publish", "sender", sender, "message", message)
	for name, ch := range b.subs {
		if name == sender {
			continue
		}
		// non blocking send.
		select {
		case ch <- message:
		default:
			b.logger.Warn("recipient buffer full", "dest", name)
		}
	}

}
