package cli

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/kschamplin/gotelem"
	"github.com/kschamplin/gotelem/socketcan"
	"github.com/kschamplin/gotelem/xbee"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

var serveFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "device",
		Aliases: []string{"d"},
		Usage:   "The XBee to connect to",
		EnvVars: []string{"XBEE_DEVICE"},
	},
	&cli.StringFlag{
		Name:    "logfile",
		Aliases: []string{"l"},
		Value:   "log.txt",
		Usage:   "file to store log to",
	},
}

var serveCmd = &cli.Command{
	Name:    "serve",
	Aliases: []string{"server", "s"},
	Usage:   "Start a telemetry server",
	Flags:   serveFlags,
	Action:  serve,
}

// FIXME: naming
// this is a server handler for i.e tcp socket, http server, socketCAN, xbee,
// etc. we can register them in init() functions.
type testThing func(cCtx *cli.Context, broker *gotelem.Broker) (err error)

type service interface {
	fmt.Stringer
	Start(cCtx *cli.Context, broker *gotelem.Broker) (err error)
	Status()
}

// this variable stores all the hanlders. It has some basic ones, but also
// can be extended on certain platforms (see cli/socketcan.go)
// or if certain features are present (see sqlite.go)
var serveThings = []testThing{}

func serve(cCtx *cli.Context) error {
	// TODO: output both to stderr and a file.
	logger := slog.New(slog.NewTextHandler(os.Stderr))

	slog.SetDefault(logger)
	broker := gotelem.NewBroker(3)

	done := make(chan struct{})
	// start the can listener
	// can logger.
	go CanDump(broker, logger.WithGroup("candump"), done)

	if cCtx.String("device") != "" {
		logger.Info("using xbee device")
		transport, err := xbee.ParseDeviceString(cCtx.String("device"))
		if err != nil {
			logger.Error("failed to open device string", "err", err)
			os.Exit(1)
		}
		go XBeeSend(broker, logger.WithGroup("xbee"), done, transport)
	}

	if cCtx.String("can") != "" {
		logger.Info("using can device")
		go canHandler(broker, logger.With("device", cCtx.String("can")), done, cCtx.String("can"))

		if strings.HasPrefix(cCtx.String("can"), "v") {
			go vcanTest(cCtx.String("can"))
		}
	}

	go broker.Start()

	// tcp listener server.
	ln, err := net.Listen("tcp", ":8082")
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
	}
	logger.Info("TCP listener started", "addr", ln.Addr().String())

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("error accepting: %v\n", err)
		}
		go handleCon(conn, broker, logger.WithGroup("tcp"), done)
	}
}


func tcpSvc(ctx *cli.Context, broker *gotelem.Broker) error {
	// TODO: extract port/ip from cli context.
	ln, err := net.Listen("tcp", ":8082")
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Printf("error accepting: %v\n", err)
		}
		go handleCon(conn, broker, slog.Default().WithGroup("tcp"), ctx.Done())
	}
}

func handleCon(conn net.Conn, broker *gotelem.Broker, l *slog.Logger, done <-chan struct{}) {
	//	reader := msgp.NewReader(conn)

	subname := fmt.Sprint("tcp", conn.RemoteAddr().String())

	l.Info("started handling", "name", subname)

	rxCh := broker.Subscribe(subname)
	defer broker.Unsubscribe(subname)
	defer conn.Close()

	for {
		select {
		case msg := <-rxCh:
			l.Info("got packet")
			// FIXME: poorly optimized
			buf := make([]byte, 0)
			buf = binary.BigEndian.AppendUint32(buf, msg.Id)
			buf = append(buf, msg.Data...)

			_, err := conn.Write(buf)
			if err != nil {
				l.Error("error writing tcp packet", "err", err)
				return
			}
		case <-done:
			return

		}
	}
}

// this spins up a new can socket on vcan0 and broadcasts a packet every second. for testing.
func vcanTest(devname string) {
	sock, err := socketcan.NewCanSocket(devname)
	if err != nil {
		slog.Error("error opening socket", "err", err)
		return
	}
	testFrame := &gotelem.Frame{
		Id:   0x234,
		Kind: gotelem.CanSFFFrame,
		Data: []byte{0, 1, 2, 3, 4, 5, 6, 7},
	}
	for {

		slog.Info("sending test packet")
		sock.Send(testFrame)
		time.Sleep(1 * time.Second)
	}
}

// connects the broker to a socket can
func canHandler(broker *gotelem.Broker, l *slog.Logger, done <-chan struct{}, devname string) {
	rxCh := broker.Subscribe("socketcan")
	sock, err := socketcan.NewCanSocket(devname)
	if err != nil {
		l.Error("error opening socket", "err", err)
		return
	}

	// start a simple dispatcher that just relays can frames.
	rxCan := make(chan gotelem.Frame)
	go func() {
		for {
			pkt, err := sock.Recv()
			if err != nil {
				l.Warn("error reading SocketCAN", "err", err)
				return
			}
			rxCan <- *pkt
		}
	}()
	for {
		select {
		case msg := <-rxCh:
			l.Info("Sending a CAN bus message", "id", msg.Id, "data", msg.Data)
			sock.Send(&msg)
		case msg := <-rxCan:
			l.Info("Got a CAN bus message", "id", msg.Id, "data", msg.Data)
			broker.Publish("socketcan", msg)
		case <-done:
			sock.Close()
			return
		}
	}
}

func CanDump(broker *gotelem.Broker, l *slog.Logger, done <-chan struct{}) {
	rxCh := broker.Subscribe("candump")
	t := time.Now()
	fname := fmt.Sprintf("candump_%d-%02d-%02dT%02d.%02d.%02d.txt",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())

	cw, err := gotelem.OpenCanWriter(fname)
	if err != nil {
		slog.Error("error opening file", "err", err)
	}

	for {
		select {
		case msg := <-rxCh:

			cw.Send(&msg)
		case <-done:
			cw.Close()
			return
		}
	}
}

func XBeeSend(broker *gotelem.Broker, l *slog.Logger, done <-chan struct{}, trspt *xbee.Transport) {
	rxCh := broker.Subscribe("xbee")
	l.Info("starting xbee send routine")

	xb, err := xbee.NewSession(trspt, l.With("device", trspt.Type()))
	if err != nil {
		l.Error("failed to start xbee session", "err", err)
		return
	}

	l.Info("connected to local xbee", "addr", xb.LocalAddr())

	for {
		select {
		case <-done:
			xb.Close()
			return
		case msg := <-rxCh:
			// TODO: take can message and send it over CAN.
			l.Info("got msg", "msg", msg)
			buf := make([]byte, 0)

			buf = binary.BigEndian.AppendUint32(buf, msg.Id)
			buf = append(buf, msg.Data...)

			_, err := xb.Write(buf)
			if err != nil {
				l.Warn("error writing to xbee", "err", err)
			}

		}
	}
}
