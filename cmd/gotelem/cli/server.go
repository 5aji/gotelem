package cli

import (
	"fmt"
	"net"
	"time"

	"github.com/kschamplin/gotelem"
	"github.com/kschamplin/gotelem/socketcan"
	"github.com/urfave/cli/v2"
)

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

	broker := gotelem.NewBroker(3)
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

func handleCon(conn net.Conn, broker *gotelem.Broker) {
	//	reader := msgp.NewReader(conn)

	conn.Close()
}

func xbeeSvc(b *gotelem.Broker) {

}

// this spins up a new can socket on vcan0 and broadcasts a packet every second. for testing.
func vcanTest() {
	sock, _ := socketcan.NewCanSocket("vcan0")
	testFrame := &gotelem.Frame{
		Id:   0x234,
		Kind: gotelem.CanSFFFrame,
		Data: []byte{0, 1, 2, 3, 4, 5, 6, 7},
	}
	for {

		fmt.Printf("sending test packet\n")
		sock.Send(testFrame)
		time.Sleep(1 * time.Second)
	}
}

func canHandler(broker *gotelem.Broker) {
	rxCh := broker.Subscribe("socketcan")
	sock, _ := socketcan.NewCanSocket("vcan0")

	// start a simple dispatcher that just relays can frames.
	rxCan := make(chan gotelem.Frame)
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
