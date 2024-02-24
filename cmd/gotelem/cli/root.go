package cli

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime/pprof"

	"github.com/urfave/cli/v2"
)

var subCmds = []*cli.Command{
	serveCmd,
	xbeeCmd,
}

var f os.File

func Execute() {
	app := &cli.App{
		Name:  "gotelem",
		Usage: "The Ultimate Telemetry Tool!",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "profile",
				Usage: "enable profiling",
			},
		},
		Before: func(ctx *cli.Context) error {
			if ctx.Bool("profile") {
				f, err := os.Create("cpuprofile")
				if err != nil {
					return err
				}
				pprof.StartCPUProfile(f)
			}
			return nil
		},
		After: func(ctx *cli.Context) error {
			if ctx.Bool("profile") {
				pprof.StopCPUProfile()
			}
			return nil
		},
		Commands: subCmds,
	}

	// setup context for cancellation.
	ctx := context.Background()
	ctx, _ = signal.NotifyContext(ctx, os.Interrupt)

	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
