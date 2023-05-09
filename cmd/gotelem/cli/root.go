package cli

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func Execute() {
	app := &cli.App{
		Name:  "gotelem",
		Usage: "see everything",
		Commands: []*cli.Command{
			serveCmd,
			xbeeCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}