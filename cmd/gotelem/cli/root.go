package cli

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

var subCmds = []*cli.Command{
	serveCmd,
	xbeeCmd,
}


func Execute() {
	app := &cli.App{
		Name:  "gotelem",
		Usage: "see everything",
		Commands: subCmds,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
