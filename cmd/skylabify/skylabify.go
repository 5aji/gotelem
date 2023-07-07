package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kschamplin/gotelem/skylab"
	"github.com/urfave/cli/v2"
	"golang.org/x/exp/slog"
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

	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:    "verbose",
			Aliases: []string{"v"},
		},
	}

	app.Action = run
	app.HideHelp = true
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

	canDumpReader := bufio.NewReader(istream)

	for {
		// dumpline looks like this:
		// (1684538768.521889) can0 200#8D643546
		dumpLine, err := canDumpReader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		// remove trailing newline
		dumpLine = strings.TrimSpace(dumpLine)

		segments := strings.Split(dumpLine, " ")

		var cd skylab.BusEvent
		var unixSeconds, unixMicros int64
		fmt.Sscanf(segments[0], "(%d.%d)", &unixSeconds, &unixMicros)
		cd.Timestamp = time.Unix(unixSeconds, unixMicros*1000) // the canlog does usec precision for the decimal part.

		// this is for the latter part, we need to split id/data
		hexes := strings.Split(segments[2], "#")

		// get the id
		id, err := strconv.ParseUint(hexes[0], 16, 64)
		if err != nil {
			return err
		}
		cd.Id = uint32(id)

		// get the data to a []byte
		rawData, err := hex.DecodeString(hexes[1])
		if err != nil {
			return err
		}

		// parse the data []byte to a skylab packet
		cd.Data, err = skylab.FromCanFrame(uint32(cd.Id), rawData)
		var idErr *skylab.UnknownIdError
		if errors.As(err, &idErr) {
			// unknown id
			slog.Info("unknown id", "err", err)
			continue
		} else if err != nil {
			return err
		}

		// format and print out the JSON.
		out, _ := json.Marshal(&cd)
		fmt.Println(string(out))

	}
}
