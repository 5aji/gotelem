package cli

import (
	"fmt"

	"github.com/kschamplin/gotelem"
	"github.com/urfave/cli/v2"
)

func init() {
	subCmds = append(subCmds, clientCmd)
}

var clientCmd = &cli.Command{
	Name:      "client",
	Aliases:   []string{"c"},
	Usage:     "interact with a gotelem server",
	ArgsUsage: "[server url]",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:    "gui",
			Aliases: []string{"g"},
			Usage:   "start a local TUI",
		},
	},
	Description: `
Connects to a gotelem server or relay. Can be used to 
	`,
	Action: client,
}


func client(ctx *cli.Context) error {
	return nil
}

