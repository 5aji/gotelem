package gotelem

import (
	"errors"
	"fmt"
	"sync"
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

func NewBroker(bufsize int) *Broker {
	b := &Broker{
		subs:      make(map[string]chan Frame),
		publishCh: make(chan BrokerRequest, 3),
		subsCh:    make(chan BrokerClient, 3),
		unsubCh:   make(chan BrokerClient, 3),
	}
	return b
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
	subs map[string] chan CANDumpJSON // contains the channel for each subsciber

	lock sync.RWMutex
}

func (b *JBroker) Subscribe(name string) (ch chan CANDumpJSON, err error) {
	// get rw lock.
	b.lock.Lock()
	defer b.lock.Unlock()
	_, ok := b.subs[name]
	if ok {
		return nil, errors.New("name already in use")
	}
	ch = make(chan CANDumpJSON, 10)

	return
}

func (b *JBroker) Unsubscribe(name string) {
	// if the channel is in use, close it, else do nothing.
	b.lock.Lock()
	defer b.lock.Unlock()
	ch, ok := b.subs[name]
	if ok {
		close(ch)
	}
	delete(b.subs, name)
}

func (b *JBroker) Publish(sender string, message CANDumpJSON) {
	go func() {
		b.lock.RLock()
		defer b.lock.RUnlock()
		for name, ch := range b.subs {
			if name == sender {
				continue
			}
			// non blocking send.
			select {
			case ch <- message:
			default:
			}
		}

	}()
}
