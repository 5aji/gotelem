//go:build linux

package cli

import (
	"github.com/kschamplin/gotelem"
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
	DefaultText: "vcan0",
}

// this function sets up the `serve` flags and services that use socketCAN
func init() {
	serveFlags = append(serveFlags, &cli.BoolFlag{Name: "test", Usage: "use vcan0 test"})
	serveFlags = append(serveFlags, canDevFlag)
	// add services for server

	serveThings = append(serveThings, socketCANService)

	// add can subcommand/actions
	// TODO: make socketcan utility commands.
}


// FIXME: add logging back in since it's missing rn

func socketCANService(cCtx *cli.Context, broker *gotelem.Broker) (err error) {
	rxCh := broker.Subscribe("socketCAN")
	sock, err := socketcan.NewCanSocket(cCtx.String("can"))
	if err != nil {
		return
	}

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
			broker.Publish("socketCAN", msg)
		case <-cCtx.Done():
			return
		}
	}

}


var socketCANCmd = &cli.Command{
	Name: "can",
	Usage: "SocketCAN utilities",
	Description: `
Various helper utilties for CAN bus on sockets. 

	`,
	Flags: []cli.Flag{
		canDevFlag,
	},

	Subcommands: []*cli.Command{
		{
			Name: "dump",
			Usage: "dump CAN packets to stdout",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name: "decode",
					Usage: "unpack known CAN packets to JSON using skylab files",
					Value: false,
				},
			},

		},
	},
}


