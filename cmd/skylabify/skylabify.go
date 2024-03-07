package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"log/slog"

	"github.com/kschamplin/gotelem/internal/logparsers"
	"github.com/kschamplin/gotelem/skylab"
	"github.com/urfave/cli/v2"
)

// this command can be used to decode candump logs and dump json output.

func main() {

	app := cli.NewApp()
	app.Name = "skylabify"
	app.Usage = "decode skylab packets"
	app.ArgsUsage = "<input file>"
	app.Commands = nil
	app.Description = `skylabify can read in candump logs and output newline-delimited JSON.
It is designed to make reading candumps fast and easy.

skylabify can be combined with jq and candump to allow for advanced queries.

Examples:
	skylabify candump.txt

	candump -L can0 | skylabify -

	skylabify previous_candump.txt | jq <some json query>

I highly suggest reading the manpages for candump and jq.  The -L option is
required for piping candump into skylabify. Likewise, data should be stored with
-l.

`
	parsersString := func() string {
		// create a string like "'telem', 'candump', 'anotherparser'"
		keys := make([]string, len(logparsers.ParsersMap))
		i := 0
		for k := range logparsers.ParsersMap {
			keys[i] = k
			i++
		}
		s := strings.Join(keys, "', '")
		return "'" + s + "'"
	}()

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
		},
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Usage:   "the format of the incoming data. One of " + parsersString,
		},
	}

	app.Action = run
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

func run(ctx *cli.Context) (err error) {
	path := ctx.Args().Get(0)
	if path == "" {
		fmt.Println("missing input file")
		cli.ShowAppHelpAndExit(ctx, int(syscall.EINVAL))
	}

	var istream *os.File
	if path == "-" {
		istream = os.Stdin
	} else {
		istream, err = os.Open(path)
		if err != nil {
			return
		}
	}

	fileReader := bufio.NewReader(istream)

	var pfun logparsers.BusEventParser

	pfun, ok := logparsers.ParsersMap[ctx.String("format")]
	if !ok {
		fmt.Println("invalid format!")
		cli.ShowAppHelpAndExit(ctx, int(syscall.EINVAL))
	}

	n_err := 0
	unknown_packets := 0

	for {
		line, err := fileReader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err // i/o failures are fatal
		}
		f, err := pfun(line)
		var idErr *skylab.UnknownIdError
		if errors.As(err, &idErr) {
			// unknown id
			slog.Info("unknown id", "err", err)
			unknown_packets++
			continue
		} else if err != nil {
			// TODO: we should consider absorbing all errors.
			slog.Error("got an error", "err", err)
			n_err++
			continue
		}

		// format and print out the JSON.
		out, _ := json.Marshal(&f)
		fmt.Println(string(out))

	}
}
