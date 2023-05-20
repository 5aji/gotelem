//go:build linux

package cli

import (
	"strings"
	"time"

	"github.com/kschamplin/gotelem"
	"github.com/kschamplin/gotelem/socketcan"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
)

// this file adds socketCAN commands and functionality when building on linux.
// It is an example of the modular architecture of the command line and server stack.

var canDevFlag = &cli.StringFlag{
	Name:        "can",
	Aliases:     []string{"c"},
	Usage:       "CAN device string",
	EnvVars:     []string{"CAN_DEVICE"},
	DefaultText: "vcan0",
}

// this function sets up the `serve` flags and services that use socketCAN
func init() {
	serveFlags = append(serveFlags, &cli.BoolFlag{Name: "test", Usage: "use vcan0 test"})
	serveFlags = append(serveFlags, canDevFlag)
	// add services for server

	serveThings = append(serveThings, &socketCANService{})

	// add can subcommand/actions
	// TODO: make socketcan utility commands.
	subCmds = append(subCmds, socketCANCmd)
}

// FIXME: add logging back in since it's missing rn

type socketCANService struct {
	sock socketcan.CanSocket
}

func (s *socketCANService) Status() {
	return
}

func (s *socketCANService) String() string {
	return ""
}

func (s *socketCANService) Start(cCtx *cli.Context, broker *gotelem.Broker, logger *slog.Logger) (err error) {
	// vcan0 demo

	if strings.HasPrefix(cCtx.String("can"), "v") {
		go vcanTest(cCtx.String("can"))
	}

	rxCh := broker.Subscribe("socketCAN")
	sock, err := socketcan.NewCanSocket(cCtx.String("can"))
	if err != nil {
		logger.Error("error opening socket", "err", err)
		return
	}

	rxCan := make(chan gotelem.Frame)

	go func() {
		for {
			pkt, err := sock.Recv()
			if err != nil {
				logger.Warn("error receiving CAN packet", "err", err)
			}
			rxCan <- *pkt
		}
	}()

	for {
		select {
		case msg := <-rxCh:
			sock.Send(&msg)
		case msg := <-rxCan:
			broker.Publish("socketCAN", msg)
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
