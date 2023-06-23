package cli

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/kschamplin/gotelem"
	"github.com/kschamplin/gotelem/xbee"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

var serveFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "xbee",
		Aliases: []string{"x"},
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

type service interface {
	fmt.Stringer
	Start(cCtx *cli.Context, broker *gotelem.JBroker, logger *slog.Logger) (err error)
	Status()
}

// this variable stores all the hanlders. It has some basic ones, but also
// can be extended on certain platforms (see cli/socketcan.go)
// or if certain features are present (see cli/sqlite.go)
var serveThings = []service{
	&xBeeService{},
	&canLoggerService{},
	&rpcService{},
	&httpService{},
}

func serve(cCtx *cli.Context) error {
	// TODO: output both to stderr and a file.
	logger := slog.New(slog.NewTextHandler(os.Stderr))

	slog.SetDefault(logger)
	broker := gotelem.NewBroker(3, logger.WithGroup("broker"))

	wg := sync.WaitGroup{}
	for _, svc := range serveThings {
		logger.Info("starting service", "svc", svc.String())
		wg.Add(1)
		go func(mySvc service, baseLogger *slog.Logger) {
			svcLogger := logger.With("svc", mySvc.String())
			defer wg.Done()
			err := mySvc.Start(cCtx, broker, svcLogger)
			if err != nil {
				logger.Error("service stopped!", "err", err, "svc", mySvc.String())
			}
		}(svc, logger)
	}

	wg.Wait()

	return nil
}

type rpcService struct {
}

func (r *rpcService) Status() {
}
func (r *rpcService) String() string {
	return "rpcService"
}

func (r *rpcService) Start(ctx *cli.Context, broker *gotelem.JBroker, logger *slog.Logger) error {
	// TODO: extract port/ip from cli context.
	ln, err := net.Listen("tcp", "0.0.0.0:8082")
	if err != nil {
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

func handleCon(conn net.Conn, broker *gotelem.JBroker, l *slog.Logger, done <-chan struct{}) {
	//	reader := msgp.NewReader(conn)

	subname := fmt.Sprint("tcp", conn.RemoteAddr().String())

	l.Info("started handling", "name", subname)
	defer conn.Close()

	rxCh, err := broker.Subscribe(subname)
	if err != nil {
		l.Error("error subscribing to connection", "err", err)
		return
	}
	defer broker.Unsubscribe(subname)

	jEncode := json.NewEncoder(conn)
	for {
		select {
		case msg := <-rxCh:
			l.Info("got packet")
			// FIXME: poorly optimized
			err := jEncode.Encode(msg)
			if err != nil {
				l.Warn("error encoding json", "err", err)
			}
		case <-done:
			return

		}
	}
}

// this spins up a new can socket on vcan0 and broadcasts a packet every second. for testing.

type canLoggerService struct {
}

func (c *canLoggerService) String() string {
	return "CanLoggerService"
}

func (c *canLoggerService) Status() {
}

func (c *canLoggerService) Start(cCtx *cli.Context, broker *gotelem.JBroker, l *slog.Logger) (err error) {
	rxCh, err := broker.Subscribe("canDump")
	if err != nil {
		return err
	}
	t := time.Now()
	fname := fmt.Sprintf("candump_%d-%02d-%02dT%02d.%02d.%02d.txt",
		t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())

	l.Info("logging to file", "filename", fname)

	f, err := os.Create(fname)
	if err != nil {
		l.Error("error opening file", "filename", fname, "err", err)
		return
	}
	enc := json.NewEncoder(f)

	for {
		select {
		case msg := <-rxCh:

			enc.Encode(msg)

		case <-cCtx.Done():
			f.Close()
			return
		}
	}
}

// xBeeService provides data over an Xbee device, either by serial or TCP
// based on the url provided in the xbee flag. see the description for details.
type xBeeService struct {
	session *xbee.Session
}

func (x *xBeeService) String() string {
	return "hello"
}
func (x *xBeeService) Status() {
}

func (x *xBeeService) Start(cCtx *cli.Context, broker *gotelem.JBroker, logger *slog.Logger) (err error) {
	if cCtx.String("xbee") == "" {
		logger.Info("not using xbee")
		return
	}
	transport, err := xbee.ParseDeviceString(cCtx.String("xbee"))
	if err != nil {
		logger.Error("failed to open xbee string", "err", err)
		return
	}
	logger.Info("using xbee device", "transport", transport)
	rxCh, err := broker.Subscribe("xbee")
	if err != nil {
		logger.Error("failed to subscribe to broker", "err", err)
	}

	x.session, err = xbee.NewSession(transport, logger.With("device", transport.Type()))
	if err != nil {
		logger.Error("failed to start xbee session", "err", err)
		return
	}
	logger.Info("connected to local xbee", "addr", x.session.LocalAddr())

	encode := json.NewEncoder(x.session)
	for {
		select {
		case <-cCtx.Done():
			x.session.Close()
			return
		case msg := <-rxCh:
			logger.Info("got msg", "msg", msg)
			encode.Encode(msg)
			if err != nil {
				logger.Warn("error writing to xbee", "err", err)
			}
		}

	}
}

type httpService struct {
}

func (h *httpService) String() string {
	return "HttpService"
}

func (h *httpService) Status() {

}

func (h *httpService) Start(cCtx *cli.Context, broker *gotelem.JBroker, logger *slog.Logger) (err error) {

	r := gotelem.TelemRouter(logger)

	http.ListenAndServe(":8080", r)
	return
}
