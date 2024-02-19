package cli

// this file contains xbee utilities.
// we can do network discovery and netcat-like things.

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"syscall"

	"log/slog"

	"github.com/kschamplin/gotelem/xbee"
	"github.com/urfave/cli/v2"
)

// context key stuff to prevent collisions
type ctxKey int

const (
	keyIODevice ctxKey = iota
)

var xbeeDeviceFlag = &cli.StringFlag{
	Name:     "device",
	Aliases:  []string{"d"},
	Usage:    "The XBee to connect to",
	Required: true,
	EnvVars:  []string{"XBEE_DEVICE"},
}

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
		xbeeDeviceFlag,
	},
	// this parses the device string and creates the io device.
	// TODO: should we create the session here instead?
	Before: func(ctx *cli.Context) error {
		transport, err := xbee.ParseDeviceString(ctx.String("device"))
		if err != nil {
			return err
		}

		ctx.Context = context.WithValue(ctx.Context, keyIODevice, transport)
		return nil
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
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name: "hex",
				},
			},
		},
	},
}

func xbeeInfo(ctx *cli.Context) error {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	transport := ctx.Context.Value(keyIODevice).(*xbee.Transport)
	xb, err := xbee.NewSession(transport, logger.With("device", transport.Type()))
	if err != nil {
		return cli.Exit(err, 1)
	}

	b, err := xb.ATCommand([2]byte{'I', 'D'}, nil, false)
	if err != nil {
		return cli.Exit(err, 1)
	}
	fmt.Printf("Network ID: %X\n", binary.BigEndian.Uint16(b))
	return nil
}

func netcat(ctx *cli.Context) error {
	if ctx.Args().Len() < 1 {

		cli.ShowSubcommandHelp(ctx)

		return cli.Exit("missing [addr] argument", int(syscall.EINVAL))

	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	transport := ctx.Context.Value(keyIODevice).(*xbee.Transport)
	xb, _ := xbee.NewSession(transport, logger.With("devtype", transport.Type()))

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

	n := <-sent

	fmt.Printf("Sent %d\n", n)

	return nil
}
