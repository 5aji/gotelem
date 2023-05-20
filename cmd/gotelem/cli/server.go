package cli

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/kschamplin/gotelem"
	"github.com/kschamplin/gotelem/xbee"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

var serveFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "device",
		Aliases: []string{"d"},
		Usage:   "The XBee to connect to. Leave blank to not use XBee",
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
type testThing func(cCtx *cli.Context, broker *gotelem.Broker, logger *slog.Logger) (err error)

type service interface {
	fmt.Stringer
	Start(cCtx *cli.Context, broker *gotelem.Broker, logger *slog.Logger) (err error)
	Status()
}

// this variable stores all the hanlders. It has some basic ones, but also
// can be extended on certain platforms (see cli/socketcan.go)
// or if certain features are present (see cli/sqlite.go)
var serveThings = []service{
	&XBeeService{},
	&CanLoggerService{},
}


func deriveLogger (oldLogger *slog.Logger, svc service) (newLogger *slog.Logger) {
	newLogger = oldLogger.With("svc", svc.String())
	return
}

func serve(cCtx *cli.Context) error {
	// TODO: output both to stderr and a file.
	logger := slog.New(slog.NewTextHandler(os.Stderr))

	slog.SetDefault(logger)
	broker := gotelem.NewBroker(3)

	done := make(chan struct{})
	// start the can listener

	go broker.Start()


	for _, svc := range serveThings {
		svcLogger := deriveLogger(logger, svc)
		logger.Info("starting service", "svc", svc.String())
		go func(mySvc service) {
			err := mySvc.Start(cCtx, broker, svcLogger)
			if err != nil {
				logger.Error("service stopped!", "err", err, "svc", mySvc.String())
			}
		}(svc)
	}

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


type rpcService struct {
}

func tcpSvc(ctx *cli.Context, broker *gotelem.Broker, logger *slog.Logger) error {
	// TODO: extract port/ip from cli context.
	ln, err := net.Listen("tcp", ":8082")
	if err != nil {
		fmt.Printf("Error listening: %v\n", err)
		logger.Warn("error listening", "err", err)
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			logger.Warn("error accepting connection", "err", err)
		}
		go handleCon(conn, broker, logger.With("addr", conn.RemoteAddr()), ctx.Done())
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

type CanLoggerService struct {
	cw gotelem.CanWriter
}

func (c *CanLoggerService) String() string {
	return "CanLoggerService"
}

func (c *CanLoggerService) Status() {
}


func (c *CanLoggerService)  Start(cCtx *cli.Context, broker *gotelem.Broker, l *slog.Logger) (err error) {
	rxCh := broker.Subscribe("candump")
	t := time.Now()
	fname := fmt.Sprintf("candump_%d-%02d-%02dT%02d.%02d.%02d.txt",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	
	l.Info("logging to file", "filename", fname)

	cw, err := gotelem.OpenCanWriter(fname)
	if err != nil {
		l.Error("error opening file", "filename", fname, "err", err)
		return
	}

	for {
		select {
		case msg := <-rxCh:

			cw.Send(&msg)
		case <-cCtx.Done():
			cw.Close()
			return
		}
	}
}

type XBeeService struct {
	session *xbee.Session
}

func (x *XBeeService) String() string {
	return "hello"
}
func (x *XBeeService) Status() {
}


func (x *XBeeService) Start(cCtx *cli.Context, broker *gotelem.Broker, logger *slog.Logger) (err error) {
	if cCtx.String("xbee") == "" {
		logger.Info("not using xbee")
		return
	}
	transport, err := xbee.ParseDeviceString(cCtx.String("device"))
	if err != nil {
		logger.Error("failed to open xbee string", "err", err)
		return
	}
	logger.Info("using xbee device", "transport", transport)
	rxCh := broker.Subscribe("xbee")

	x.session, err = xbee.NewSession(transport, logger.With("device", transport.Type()))
	if err != nil {
		logger.Error("failed to start xbee session", "err", err)
		return
	}
	logger.Info("connected to local xbee", "addr", x.session.LocalAddr())

	for {
		select {
		case <-cCtx.Done():
			x.session.Close()
			return
		case msg := <-rxCh:
			logger.Info("got msg", "msg", msg)
			buf := make([]byte, 0)

			buf = binary.BigEndian.AppendUint32(buf, msg.Id)
			buf = append(buf, msg.Data...)

			_, err := x.session.Write(buf)
			if err != nil {
				logger.Warn("error writing to xbee", "err", err)
			}
		}


	}
}

