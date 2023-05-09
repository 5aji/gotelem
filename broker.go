package gotelem

import "fmt"

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

// TODO: don't use channels for everything to avoid using a mutex
