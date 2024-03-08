//go:build linux

package cli

import (
	"errors"
	"io"
	"time"

	"github.com/kschamplin/gotelem/internal/can"
	"github.com/kschamplin/gotelem/skylab"
	"github.com/kschamplin/gotelem/socketcan"
	"github.com/urfave/cli/v2"
)

// this file adds socketCAN commands and functionality when building on linux.
// It is an example of the modular architecture of the command line and server stack.

var canDevFlag = &cli.StringFlag{
	Name:    "can",
	Aliases: []string{"c"},
	Usage:   "CAN device string",
	EnvVars: []string{"CAN_DEVICE"},
}

// this function sets up the `serve` flags and services that use socketCAN
func init() {
	// add the CAN flags to the serve command
	serveCmd.Flags = append(serveCmd.Flags, &cli.BoolFlag{Name: "test", Usage: "use vcan0 test"})
	serveCmd.Flags = append(serveCmd.Flags, canDevFlag)

	// add services for server
	serveThings = append(serveThings, &socketCANService{})

	// add can subcommand/actions
	// TODO: make more utility commands.
	subCmds = append(subCmds, socketCANCmd)
}

type socketCANService struct {
	name string
	sock *socketcan.CanSocket
}

func (s *socketCANService) Status() {
}

func (s *socketCANService) String() string {
	if s.name == "" {
		return "socketCAN"
	}
	return s.name
}

// Start starts the socketCAN service - emitting packets sent from the broker.
func (s *socketCANService) Start(cCtx *cli.Context, deps svcDeps) (err error) {

	logger := deps.Logger
	broker := deps.Broker
	tdb := deps.Db

	if !cCtx.IsSet("can") {
		logger.Debug("no can device provided, skip")
		return
	}

	s.sock, err = socketcan.NewCanSocket(cCtx.String("can"))
	if err != nil {
		logger.Error("error opening socket", "err", err)
		return
	}
	defer s.sock.Close()
	s.name = s.sock.Name()

	// connect to the broker
	rxCh, err := broker.Subscribe("socketCAN")
	if err != nil {
		return err
	}
	defer broker.Unsubscribe("socketCAN")

	// make a channel to receive socketCAN frames.
	rxCan := make(chan can.Frame)

	go func() {
		for {
			pkt, err := s.sock.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				logger.Warn("error receiving CAN packet", "err", err)
			}
			rxCan <- *pkt
		}
	}()

	var frame can.Frame
	for {
		select {
		case msg := <-rxCh:

			frame, err = skylab.ToCanFrame(msg.Data)
			if err != nil {
				logger.Warn("error encoding can frame", "name", msg.Name, "err", err)
				continue
			}

			s.sock.Send(&frame)

		case msg := <-rxCan:
			p, err := skylab.FromCanFrame(msg)
			if err != nil {
				logger.Warn("error parsing can packet", "id", msg.Id, "err", err)
				continue
			}
			event := skylab.BusEvent{
				Timestamp: time.Now(),
				Name:      p.String(),
				Data:      p,
			}
			broker.Publish("socketCAN", event)
			tdb.AddEventsCtx(cCtx.Context, event)
		case <-cCtx.Done():
			// close the socket.
			s.sock.Close()
			return
		}
	}
}

var socketCANCmd = &cli.Command{
	Name:  "can",
	Usage: "SocketCAN utilities",
	Description: `
Various helper utilties for CAN bus on sockets.
	`,
	Flags: []cli.Flag{
		canDevFlag,
	},

	Subcommands: []*cli.Command{
		{
			Name:  "dump",
			Usage: "dump CAN packets to stdout",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "decode",
					Usage: "unpack known CAN packets to JSON using skylab files",
					Value: false,
				},
			},
		},
	},
}
