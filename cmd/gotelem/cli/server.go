package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"math"
	"time"
	"os"
	"sync"

	"log/slog"

	"github.com/kschamplin/gotelem"
	"github.com/kschamplin/gotelem/skylab"
	"github.com/kschamplin/gotelem/xbee"
	"github.com/urfave/cli/v2"
)

var serveFlags = []cli.Flag{
	&cli.StringFlag{
		Name:    "xbee",
		Aliases: []string{"x"},
		Usage:   "The XBee to connect to. Leave blank to not use XBee",
		EnvVars: []string{"XBEE_DEVICE"},
	},
	&cli.PathFlag{
		Name:        "logfile",
		Aliases:     []string{"l"},
		DefaultText: "log.txt",
		Usage:       "file to store log to",
	},
	&cli.PathFlag{
		Name:        "db",
		Aliases:     []string{"d"},
		DefaultText: "gotelem.db",
		Usage:       "database to serve, if not specified will use memory",
	},
	&cli.BoolFlag{
		Name: "demo",
		Usage: "enable the demo packet stream",
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
	Start(cCtx *cli.Context, deps svcDeps) (err error)
}

type svcDeps struct {
	Broker *gotelem.Broker
	Db     *gotelem.TelemDb
	Logger *slog.Logger
}

// this variable stores all the hanlders. It has some basic ones, but also
// can be extended on certain platforms (see cli/socketcan.go)
// or if certain features are present (see cli/sqlite.go)
var serveThings = []service{
	&xBeeService{},
	&httpService{},
	&DemoService{},
}

func serve(cCtx *cli.Context) error {
	// TODO: output both to stderr and a file.
	var output io.Writer = os.Stderr

	if cCtx.IsSet("logfile") {
		// open the file.
		p := cCtx.Path("logfile")
		f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		output = io.MultiWriter(os.Stderr, f)
	}
	// create a new logger
	logger := slog.New(slog.NewTextHandler(output, nil))

	slog.SetDefault(logger)

	broker := gotelem.NewBroker(20, logger.WithGroup("broker"))

	// open database
	dbPath := "gotelem.db"
	if cCtx.IsSet("db") {
		dbPath = cCtx.Path("db")
	}
	logger.Info("opening database", "path", dbPath)
	db, err := gotelem.OpenTelemDb(dbPath)
	if err != nil {
		return err
	}

	wg := sync.WaitGroup{}

	deps := svcDeps{
		Logger: logger,
		Broker: broker,
		Db:     db,
	}

	for _, svc := range serveThings {
		logger.Info("starting service", "service", svc.String())
		wg.Add(1)
		go func(mySvc service, baseLogger *slog.Logger) {
			svcLogger := logger.With("service", mySvc.String())
			s := deps
			s.Logger = svcLogger
			defer wg.Done()
			// TODO: recover
			err := mySvc.Start(cCtx, s)
			if err != nil {
				logger.Error("service stopped!", "err", err, "service", mySvc.String())
			}
		}(svc, logger)
	}

	wg.Wait()

	return nil
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

func (x *xBeeService) Start(cCtx *cli.Context, deps svcDeps) (err error) {
	logger := deps.Logger
	broker := deps.Broker
	tdb := deps.Db
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

	// these are the ways we send/recieve data. we could swap for binary format
	// TODO: buffering and/or binary encoding instead of json which is horribly ineffective.
	xbeeTxer := json.NewEncoder(x.session)
	xbeeRxer := json.NewDecoder(x.session)

	go func() {
		for {
			var p skylab.BusEvent
			err := xbeeRxer.Decode(&p)
			if err != nil {
				logger.Error("failed to decode xbee packet")
			}
			broker.Publish("xbee", p)
			tdb.AddEventsCtx(cCtx.Context, p)
		}
	}()
	for {
		select {
		case <-cCtx.Done():
			x.session.Close()
			return
		case msg := <-rxCh:
			logger.Info("got msg", "msg", msg)
			err := xbeeTxer.Encode(msg)
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

func (h *httpService) Start(cCtx *cli.Context, deps svcDeps) (err error) {

	logger := deps.Logger
	broker := deps.Broker
	db := deps.Db

	r := gotelem.TelemRouter(logger, broker, db)

	//

	/// TODO: use custom port if specified
	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
	go func() {
		<-cCtx.Done()
		logger.Info("shutting down server")
		server.Shutdown(cCtx.Context)
	}()
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		logger.ErrorContext(cCtx.Context, "Error listening", "err", err)
	}
	return
}


type DemoService struct {
}

func (d *DemoService) String() string {
	return "demo service"
}

func (d *DemoService) Start(cCtx *cli.Context, deps svcDeps) (err error) {
	if !cCtx.Bool("demo") {
		return 
	}

	broker := deps.Broker
	bmsPkt := &skylab.BmsMeasurement{
		Current: 1.23,
		BatteryVoltage: 11111,
		AuxVoltage: 22222,
	}
	wslPkt := &skylab.WslVelocity{
		MotorVelocity: 0,
		VehicleVelocity: 100.0,
	}
	var next skylab.Packet = bmsPkt
	for {
		select {
		case <-cCtx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			// send the next packet.
			if next == bmsPkt {
				bmsPkt.Current = float32(math.Sin(float64(time.Now().UnixMilli()) / 2000.0))
				ev := skylab.BusEvent{
					Timestamp: time.Now(),
					Name: next.String(),
					Data: next,
				}
				broker.Publish("livestream", ev)
				next = wslPkt
			} else {
				// send the wsl
				ev := skylab.BusEvent{
					Timestamp: time.Now(),
					Name: next.String(),
					Data: next,
				}
				broker.Publish("livestream", ev)
				next = bmsPkt
			}



		}
	}
}
