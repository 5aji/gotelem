package main

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/kschamplin/gotelem/internal/can"
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
		&cli.StringFlag{
			Name:    "format",
			Aliases: []string{"f"},
			Usage:   "the format of the incoming data. One of 'telem', 'candump'",
		},
	}

	app.Action = run
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}

// A FormatError is an error when parsing a format. Typically we simply ignore
// these and move on, but they can optionally wrap another error that is fatal.
type FormatError struct {
	msg string
	err error
}

func (e *FormatError) Error() string {
	return fmt.Sprintf("%s:%s", e.msg, e.err.Error())
}
func (e *FormatError) Unwrap() error {
	return e.err
}

func NewFormatError(msg string, err error) error {
	return &FormatError{msg: msg, err: err}
}

// A Parser takes a string containing one line of a particular log file
// and returns an associated skylab.BusEvent representing the packet.
// if no packet is found, an error is returned instead.
type ParserFunc func(string) (skylab.BusEvent, error)

func parseCanDumpLine(dumpLine string) (b skylab.BusEvent, err error) {
	// dumpline looks like this:
	// (1684538768.521889) can0 200#8D643546
	// remove trailing newline
	dumpLine = strings.TrimSpace(dumpLine)
	segments := strings.Split(dumpLine, " ")

	var unixSeconds, unixMicros int64
	fmt.Sscanf(segments[0], "(%d.%d)", &unixSeconds, &unixMicros)
	b.Timestamp = time.Unix(unixSeconds, unixMicros)

	// now we extract the remaining data:
	hexes := strings.Split(segments[2], "#") // first portion is id, second is data

	id, err := strconv.ParseUint(hexes[0], 16, 64)
	if err != nil {
		err = NewFormatError("failed to parse id", err)
		return
	}
	if (len(hexes[1]) % 2) != 0 {
		err = NewFormatError("odd number of hex characters", nil)
		return
	}
	rawData, err := hex.DecodeString(hexes[1])
	if err != nil {
		err = NewFormatError("failed to decode hex data", err)
		return
	}

	frame := can.Frame{
		// TODO: fix extended ids. we assume not extended for now.
		Id:   can.CanID{Id: uint32(id), Extended: false},
		Data: rawData,
		Kind: can.CanDataFrame,
	}

	b.Data, err = skylab.FromCanFrame(frame)

	if err != nil {
		err = NewFormatError("failed to parse can frame", err)
		return
	}

	// set the name
	b.Name = b.Data.String()

	return
}

func parseTelemLogLine(line string) (b skylab.BusEvent, err error) {
	// strip trailng newline since we rely on it being gone
	line = strings.TrimSpace(line)
	// data is of the form
	// 1698180835.318 0619D80564080EBE241
	// the second part there is 3 nibbles (12 bits, 3 hex chars) for can ID,
	// the rest is data.
	// this regex does the processing.
	r := regexp.MustCompile(`^(\d+).(\d{3}) (\w{3})(\w+)$`)

	// these files tend to get corrupted. there are all kinds of nasties that can happen.
	// defense against random panics
	defer func() {
		if r := recover(); r != nil {
			err = NewFormatError("caught panic", nil)
		}
	}()
	a := r.FindStringSubmatch(line)
	if a == nil {
		err = NewFormatError("no regex match", nil)
		return
	}
	var unixSeconds, unixMillis int64
	// note that a contains 5 elements, the first being the full match.
	// so we start from the second element
	unixSeconds, err = strconv.ParseInt(a[1], 10, 0)
	if err != nil {
		err = NewFormatError("failed to parse unix seconds", err)
		return
	}
	unixMillis, err = strconv.ParseInt(a[2], 10, 0)
	if err != nil {
		err = NewFormatError("failed to parse unix millis", err)
		return
	}
	ts := time.Unix(unixSeconds, unixMillis*1e6)

	id, err := strconv.ParseUint(a[3], 16, 16)
	if err != nil {
		err = NewFormatError("failed to parse id", err)
		return
	}

	if len(a[4])%2 != 0 {
		// odd hex chars, protect against a panic
		err = NewFormatError("wrong amount of hex chars", nil)
	}
	rawData, err := hex.DecodeString(a[4])
	if err != nil {
		err = NewFormatError("failed to parse hex data", err)
		return
	}
	frame := can.Frame{
		Id:   can.CanID{Id: uint32(id), Extended: false},
		Data: rawData,
		Kind: can.CanDataFrame,
	}
	b.Timestamp = ts
	b.Data, err = skylab.FromCanFrame(frame)
	if err != nil {
		err = NewFormatError("failed to parse can frame", err)
		return
	}
	b.Name = b.Data.String()

	return
}

var parseMap = map[string]ParserFunc{
	"telem":   parseTelemLogLine,
	"candump": parseCanDumpLine,
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

	var pfun ParserFunc

	pfun, ok := parseMap[ctx.String("format")]
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
			fmt.Printf("got an error %v\n", err)
			n_err++
			continue
		}

		// format and print out the JSON.
		out, _ := json.Marshal(&f)
		fmt.Println(string(out))

	}
}
