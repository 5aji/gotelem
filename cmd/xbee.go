package cmd

// this file contains xbee utilities.
// we can do network discovery and netcat-like things.

import (
	"fmt"
	"io"
	"os"

	"github.com/kschamplin/gotelem/xbee"
	"github.com/urfave/cli/v2"
	"go.bug.st/serial"
	"golang.org/x/exp/slog"
)

var xbeeCmd = &cli.Command{
	Name:    "xbee",
	Aliases: []string{"x"},
	Usage:   "Utilities for XBee",
	Description: `
Allows for testing and debugging XBee networks and devices.
The "device" parameter is not optional, and can be any of the following formats:
		tcp://192.168.4.5:8430 
		COM1
		/dev/ttyUSB0:115200
For serial devices (COM1 and /dev/ttyUSB0), you can specify the baud rate
using a ':'. If excluded the baud rate will default to 9600. Note that
if using the native USB of the XLR Pro, the baud rate setting has no effect.

TCP/UDP connections require a port and will fail if one is not provided.

	`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "device",
			Aliases:  []string{"d"},
			Usage:    "The XBee to connect to",
			Required: true,
			EnvVars:  []string{"XBEE_DEVICE"},
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:            "info",
			Usage:           "get information about an xbee device",
			Action:          xbeeInfo,
			HideHelpCommand: true,
		},
		{
			Name:      "netcat",
			Aliases:   []string{"nc"},
			ArgsUsage: "[addr]",
			Usage:     "send data from stdio over the xbee",
			Description: `
netcat emulates the nc command. It reads data from stdin and transmits it to
[addr] on the XBee network. If [addr] is FFFF or not present, it will broadcast
the data to all listening devices. Data received from the network will be
writtend to stdout.
			`,
			Action:          netcat,
			HideHelpCommand: true,
		},
	},
}

func xbeeInfo(ctx *cli.Context) error {

	logger := slog.New(slog.NewTextHandler(os.Stderr))
	xb, err := xbee.NewSerialXBee("/dev/ttyACM0", &serial.Mode{}, logger)
	if err != nil {
		return cli.Exit(err, 1)
	}

	b, err := xb.ATCommand([2]rune{'I', 'D'}, nil, false)
	if err != nil {
		return cli.Exit(err, 1)
	}
	fmt.Println(b)
	return nil

}
func netcat(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {

		cli.ShowSubcommandHelp(ctx)

		return cli.Exit("missing [addr] argument", 1)

	}
	// basically create two pipes.
	logger := slog.New(slog.NewTextHandler(os.Stderr))
	xb, _ := xbee.NewSerialXBee("/dev/ttyACM0", &serial.Mode{}, logger)

	sent := make(chan int64)
	streamCopy := func(r io.ReadCloser, w io.WriteCloser) {
		defer r.Close()
		defer w.Close()
		n, err := io.Copy(w, r)
		if err != nil {
			logger.Warn("got error copying", "err", err)
		}
		sent <- n
	}

	go streamCopy(os.Stdin, xb)
	go streamCopy(xb, os.Stdout)

	<-sent

	return nil
}
