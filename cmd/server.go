package cmd

import (
	"github.com/urfave/cli/v2"
)

var serveCmd = &cli.Command{
	Name:    "serve",
	Aliases: []string{"server", "s"},
	Usage:   "Start a telemetry server",
}
