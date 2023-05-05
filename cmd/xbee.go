package cmd

// this file contains xbee utilities.
// we can do network discovery and netcat-like things.

import (
	"github.com/urfave/cli/v2"
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

TCP/UDP connections require a port.
	`,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "device",
			Aliases: []string{"d"},
			Usage:   "The XBee to connect to",
		},
	},
	Subcommands: []*cli.Command{
		{
			Name:  "info",
			Usage: "get information about an xbee device",
		},
		{
			Name:      "netcat",
			Aliases:   []string{"nc"},
			ArgsUsage: "[addr]",
			Usage:     "send data from stdio over the xbee",
		},
	},
}
