//go:build linux

package cli

import (
	"strings"
	"time"

	"github.com/kschamplin/gotelem"
	"github.com/kschamplin/gotelem/skylab"
	"github.com/kschamplin/gotelem/socketcan"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
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
	return
}

func (s *socketCANService) String() string {
	if s.name == "" {
		return "socketCAN"
	}
	return s.name
}

func (s *socketCANService) Start(cCtx *cli.Context, broker *gotelem.JBroker, logger *slog.Logger) (err error) {
	// vcan0 demo

	if cCtx.String("can") == "" {
		logger.Info("no can device provided")
		return
	}

	if strings.HasPrefix(cCtx.String("can"), "v") {
		go vcanTest(cCtx.String("can"))
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
	rxCan := make(chan gotelem.Frame)

	go func() {
		for {
			pkt, err := s.sock.Recv()
			if err != nil {
				logger.Warn("error receiving CAN packet", "err", err)
			}
			rxCan <- *pkt
		}
	}()

	var frame gotelem.Frame
	for {
		select {
		case msg := <-rxCh:

			id, d, _ := skylab.ToCanFrame(msg.Data)

			frame.Id = id
			frame.Data = d

			s.sock.Send(&frame)

		case msg := <-rxCan:
			p, err := skylab.FromCanFrame(msg.Id, msg.Data)
			if err != nil {
				logger.Warn("error parsing can packet", "id", msg.Id)
				continue
			}
			cde := skylab.BusEvent{
				Timestamp: float64(time.Now().UnixNano()) / 1e9,
				Id:        uint64(msg.Id),
				Data:      p,
			}
			broker.Publish("socketCAN", cde)
		case <-cCtx.Done():
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

func vcanTest(devname string) {
	sock, err := socketcan.NewCanSocket(devname)
	if err != nil {
		slog.Error("error opening socket", "err", err)
		return
	}
	testPkt := skylab.WslMotorCurrentVector{
		Iq: 0.1,
		Id: 0.2,
	}

	id, data, err := skylab.ToCanFrame(&testPkt)
	testFrame := gotelem.Frame{
		Id:   id,
		Data: data,
		Kind: gotelem.CanSFFFrame,
	}

	for {
		slog.Info("sending test packet")
		sock.Send(&testFrame)
		time.Sleep(1 * time.Second)
	}
}
