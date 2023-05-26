package gotelem

import (
	"errors"
	"fmt"
	"sync"

	"golang.org/x/exp/slog"
)

type BrokerRequest struct {
	Source string // the name of the sender
	Msg    Frame  // the message to send
}
type BrokerClient struct {
	Name string     // the name of the client
	Ch   chan Frame // the channel to send frames to this client
}
type Broker struct {
	subs map[string]chan Frame

	publishCh chan BrokerRequest

	subsCh  chan BrokerClient
	unsubCh chan BrokerClient
}


// Start runs the broker and sends messages to the subscribers (but not the sender)
func (b *Broker) Start() {

	for {
		select {
		case newClient := <-b.subsCh:
			b.subs[newClient.Name] = newClient.Ch
		case req := <-b.publishCh:
			for name, ch := range b.subs {
				if name == req.Source {
					continue // don't send to ourselves.
				}
				// a kinda-inelegant non-blocking push.
				// if we can't do it, we just drop it. this should ideally never happen.
				select {
				case ch <- req.Msg:
				default:
					fmt.Printf("we dropped a packet to dest %s", name)
				}
			}
		case clientToRemove := <-b.unsubCh:
			close(b.subs[clientToRemove.Name])
			delete(b.subs, clientToRemove.Name)
		}
	}
}

func (b *Broker) Publish(name string, msg Frame) {
	breq := BrokerRequest{
		Source: name,
		Msg:    msg,
	}
	b.publishCh <- breq
}

func (b *Broker) Subscribe(name string) <-chan Frame {
	ch := make(chan Frame, 3)

	bc := BrokerClient{
		Name: name,
		Ch:   ch,
	}
	b.subsCh <- bc
	return ch
}

func (b *Broker) Unsubscribe(name string) {
	bc := BrokerClient{
		Name: name,
	}
	b.unsubCh <- bc
}




type JBroker struct {
	subs map[string] chan CANDumpEntry // contains the channel for each subsciber

	logger *slog.Logger
	lock sync.RWMutex
	bufsize int // size of chan buffer in elements.
}


func NewBroker(bufsize int, logger *slog.Logger) *JBroker {
	return &JBroker{
		subs: make(map[string]chan CANDumpEntry),
		logger: logger,
		bufsize: bufsize,
	}
}

func (b *JBroker) Subscribe(name string) (ch chan CANDumpEntry, err error) {
	// get rw lock.
	b.lock.Lock()
	defer b.lock.Unlock()
	_, ok := b.subs[name]
	if ok {
		return nil, errors.New("name already in use")
	}
	b.logger.Info("new subscriber", "name", name)
	ch = make(chan CANDumpEntry, b.bufsize)

	return
}

func (b *JBroker) Unsubscribe(name string) {
	// remove the channel from the map. We don't need to close it.
	b.lock.Lock()
	defer b.lock.Unlock()
	delete(b.subs, name)
}

func (b *JBroker) Publish(sender string, message CANDumpEntry) {
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
