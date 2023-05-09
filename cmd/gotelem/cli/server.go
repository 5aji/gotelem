package cli

import (
	"fmt"
	"net"
	"time"

	"github.com/kschamplin/gotelem/can"
	"github.com/kschamplin/gotelem/socketcan"
	"github.com/urfave/cli/v2"
)

const xbeeCategory = "XBee settings"

var serveCmd = &cli.Command{
	Name:    "serve",
	Aliases: []string{"server", "s"},
	Usage:   "Start a telemetry server",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "xbee", Aliases: []string{"x"}, Usage: "Find and connect to an XBee"},
	},
	Action: func(ctx *cli.Context) error {
		serve(ctx.Bool("xbee"))
		return nil
	},
}

func serve(useXbee bool) {

	broker := NewBroker(3)
	// start the can listener
	go vcanTest()
	go canHandler(broker)
	go broker.Start()
	ln, err := net.Listen("tcp", ":8082")
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
	}
	fmt.Printf("Listening on :8082\n")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("error accepting: %v\n", err)
		}
		go handleCon(conn, broker)
	}
}

func handleCon(conn net.Conn, broker *Broker) {
	//	reader := msgp.NewReader(conn)

	conn.Close()
}

func xbeeSvc(b *Broker) {

}

// this spins up a new can socket on vcan0 and broadcasts a packet every second. for testing.
func vcanTest() {
	sock, _ := socketcan.NewCanSocket("vcan0")
	testFrame := &can.Frame{
		Id:   0x234,
		Kind: can.SFF,
		Data: []byte{0, 1, 2, 3, 4, 5, 6, 7},
	}
	for {

		fmt.Printf("sending test packet\n")
		sock.Send(testFrame)
		time.Sleep(1 * time.Second)
	}
}

func canHandler(broker *Broker) {
	rxCh := broker.Subscribe("socketcan")
	sock, _ := socketcan.NewCanSocket("vcan0")

	// start a simple dispatcher that just relays can frames.
	rxCan := make(chan can.Frame)
	go func() {
		for {
			pkt, _ := sock.Recv()
			rxCan <- *pkt
		}
	}()
	for {
		select {
		case msg := <-rxCh:
			sock.Send(&msg)
		case msg := <-rxCan:
			fmt.Printf("got a packet from the can %v\n", msg)
			broker.Publish("socketcan", msg)
		}
	}
}

type BrokerRequest struct {
	Source string    // the name of the sender
	Msg    can.Frame // the message to send
}
type BrokerClient struct {
	Name string         // the name of the client
	Ch   chan can.Frame // the channel to send frames to this client
}
type Broker struct {
	subs map[string]chan can.Frame

	publishCh chan BrokerRequest

	subsCh  chan BrokerClient
	unsubCh chan BrokerClient
}

func NewBroker(bufsize int) *Broker {
	b := &Broker{
		subs:      make(map[string]chan can.Frame),
		publishCh: make(chan BrokerRequest, 3),
		subsCh:    make(chan BrokerClient, 3),
		unsubCh:   make(chan BrokerClient, 3),
	}
	return b
}

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

func (b *Broker) Publish(name string, msg can.Frame) {
	breq := BrokerRequest{
		Source: name,
		Msg:    msg,
	}
	b.publishCh <- breq
}

func (b *Broker) Subscribe(name string) <-chan can.Frame {
	ch := make(chan can.Frame, 3)

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