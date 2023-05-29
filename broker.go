package gotelem

import (
	"errors"
	"sync"

	"github.com/kschamplin/gotelem/skylab"
	"golang.org/x/exp/slog"
)

type JBroker struct {
	subs map[string]chan skylab.BusEvent // contains the channel for each subsciber

	logger  *slog.Logger
	lock    sync.RWMutex
	bufsize int // size of chan buffer in elements.
}

func NewBroker(bufsize int, logger *slog.Logger) *JBroker {
	return &JBroker{
		subs:    make(map[string]chan skylab.BusEvent),
		logger:  logger,
		bufsize: bufsize,
	}
}

func (b *JBroker) Subscribe(name string) (ch chan skylab.BusEvent, err error) {
	// get rw lock.
	b.lock.Lock()
	defer b.lock.Unlock()
	_, ok := b.subs[name]
	if ok {
		return nil, errors.New("name already in use")
	}
	b.logger.Info("new subscriber", "name", name)
	ch = make(chan skylab.BusEvent, b.bufsize)

	b.subs[name] = ch
	return
}

func (b *JBroker) Unsubscribe(name string) {
	// remove the channel from the map. We don't need to close it.
	b.lock.Lock()
	defer b.lock.Unlock()
	delete(b.subs, name)
}

func (b *JBroker) Publish(sender string, message skylab.BusEvent) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	for name, ch := range b.subs {
		if name == sender {
			continue
		}
		// non blocking send.
		select {
		case ch <- message:
			b.logger.Debug("sent message", "dest", name, "src", sender)
		default:
			b.logger.Warn("recipient buffer full", "dest", name)
		}
	}

}
