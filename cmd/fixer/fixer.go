package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"log/slog"

	"github.com/kschamplin/gotelem/skylab"
	"github.com/urfave/cli/v2"
)

// fixer resolves four major issues with CAN dumps:
// 1. ISO8601 timestamps
// 2. Unix seconds timestamps
// 3. Missing/broken timestamps (retime)
// 4. missing names

func main() {
	app := cli.NewApp()
	app.Name = "fixer"
	app.Usage = " fix skylabify outputs"
	app.ArgsUsage = "<input file>"
	app.Description = `fixer fixes four major issues with CAN dumps
1. ISO8601 timestamps --time iso8601
2. Unix seconds timestamps --time seconds
3. Missing/broken timestamps --retime
4. missing names (enabled by default, --no-rename to skip)
	`
	app.Flags = []cli.Flag{
		&cli.BoolFlag{
			Name:  "retime",
			Usage: "ignore timestamps and retime data based on 200ms heartbeat packet",
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "Timestamp format. One of 'iso8601', 'seconds'. Ignored if using retime.",
		},
		&cli.BoolFlag{
			Name:  "no-rename",
			Usage: "skip changing packet names",
		},
		&cli.StringFlag{
			Name:  "output",
			Usage: "file to output to. defaults to stdout.",
		},
	}

	app.Before = validateArgs
	app.Action = run
	app.HideHelp = true
	app.Run(os.Args)
}

var allowedFormats = []string{
	"iso8601",
	"seconds",
}

func checkFormat(input string) bool {
	for _, allowed := range allowedFormats {
		if input == allowed {
			return true
		}
	}
	return false
}

func validateArgs(ctx *cli.Context) error {
	if ctx.IsSet("format") {
		format := ctx.String("format")
		if !checkFormat(format) {
			fmt.Printf("invalid format string, got %s, must be one of %s", format, strings.Join(allowedFormats, ", "))
			cli.ShowAppHelpAndExit(ctx, int(syscall.EINVAL))
		}

	}
	if ctx.Args().Get(0) == "" {
		fmt.Println("missing input file")
		cli.ShowAppHelpAndExit(ctx, int(syscall.EINVAL))
	}
	return nil
}

func run(ctx *cli.Context) (err error) {
	path := ctx.Args().Get(0)

	var istream *os.File
	if path == "-" {
		istream = os.Stdin
	} else {
		istream, err = os.Open(path)
		if err != nil {
			return
		}
	}
	var ostream *os.File
	oFilename := ctx.Args().Get(1)
	if oFilename == "" {
		ostream = os.Stdout
	} else {
		ostream, err = os.Create(oFilename)
		if err != nil {
			return
		}
	}

	iReader := json.NewDecoder(istream)

	shouldRetime := ctx.Bool("retime")

	// used for retime - increment by 200 ms every time we see a WsrStatusPacket
	currentTick := time.Now().UnixMilli()

	for {
		// read a line of json and fix names if we should.

		// FIXME: handle missing trailing newline.
		// based on the fomat string, we should parse the data raw...
		var res brokenCANMsg
		err := iReader.Decode(&res)

		// float64 can have issues when represeneting
		// unix milliseconds. the JSON decoder supports
		// a "Number" struct that can be either.
		iReader.UseNumber()
		if err != nil {
			return err
		}

		var goodPkt skylab.RawJsonEvent

		goodPkt.Data = res.Data
		goodPkt.Id = uint32(res.Id)
		goodPkt.Name = res.Name
		// wait to decode packet before rename.

		switch ts := res.Timestamp.(type) {
		case json.Number:
			// if it contains a decimal.
			if strings.Contains(ts.String(), ".") {
				// it's a float.
				t, _ := ts.Float64()
				t = t * 1000
				goodPkt.Timestamp = int64(t)
			} else {
				// it's an int.
				t, _ := ts.Int64()
				if ctx.String("format") == "seconds" {
					t = t * 1000
				}
				goodPkt.Timestamp = t
			}
		case string:
			// parse as ISO8601
			// use unix millis
			var t time.Time
			err := t.UnmarshalText([]byte(ts))
			if err != nil {
				panic(err)
			}
			goodPkt.Timestamp = t.UnixMilli()

		}

		if shouldRetime {
			if goodPkt.Id == uint32(skylab.WsrStatusInformationId) {
				// bump the clock 200ms
				currentTick += 200
			}
			goodPkt.Timestamp = currentTick
		}

		// now, spit it out.
		var bEv skylab.BusEvent
		bEv.Timestamp = time.UnixMilli(goodPkt.Timestamp)
		bEv.Id = goodPkt.Id
		bEv.Data, err = skylab.FromJson(goodPkt.Id, goodPkt.Data)
		var idErr *skylab.UnknownIdError
		if errors.As(err, &idErr) {
			// unknown id
			slog.Info("unknown id", "err", err)
			continue
		} else if err != nil {
			return err
		}

		out, err := json.Marshal(&bEv)
		if err != nil {
			panic(err)
		}
		fmt.Fprintln(ostream, string(out))
	}
}

type brokenCANMsg struct {
	Timestamp any `json:"ts"`
	Id        int32
	Name      string
	Data      json.RawMessage
}
