package logparsers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/kschamplin/gotelem/internal/can"
	"github.com/kschamplin/gotelem/skylab"
)

// A FormatError is an error when parsing a format. Typically we simply ignore
// these and move on, but they can optionally wrap another error that is fatal.
type FormatError struct {
	msg string
	err error
}

func (e *FormatError) Error() string {
	if e.err != nil {
		return fmt.Sprintf("%s:%s", e.msg, e.err.Error())
	}
	return e.msg

}
func (e *FormatError) Unwrap() error {
	return e.err
}

// NewFormatError constructs a new format error.
func NewFormatError(msg string, err error) error {
	return &FormatError{msg: msg, err: err}
}

// type CanFrameParser is a function that takes a string
// and returns a can frame. This is useful for common
// can dump formats.
type CanFrameParser func(string) (can.Frame, time.Time, error)

var candumpRegex = regexp.MustCompile(`^\((\d+)\.(\d{6})\) \w+ (\w+)#(\w+)$`)

func parseCanDumpLine(dumpLine string) (frame can.Frame, ts time.Time, err error) {
	frame = can.Frame{}
	ts = time.Unix(0, 0)
	// dumpline looks like this:
	// (1684538768.521889) can0 200#8D643546
	// remove trailing newline/whitespaces
	dumpLine = strings.TrimSpace(dumpLine)
	m := candumpRegex.FindStringSubmatch(dumpLine)
	if m == nil || len(m) != 5 {
		err = NewFormatError("no regex match", nil)
		return
	}

	var unixSeconds, unixMicros int64

	unixSeconds, err = strconv.ParseInt(m[1], 10, 0)
	if err != nil {
		err = NewFormatError("failed to parse unix seconds", err)
		return
	}
	unixMicros, err = strconv.ParseInt(m[2], 10, 0)
	if err != nil {
		err = NewFormatError("failed to parse unix micros", err)
		return
	}

	id, err := strconv.ParseUint(m[3], 16, 64)
	if err != nil {
		err = NewFormatError("failed to parse id", err)
		return
	}
	if (len(m[4]) % 2) != 0 {
		err = NewFormatError("odd number of hex characters", nil)
		return
	}
	rawData, err := hex.DecodeString(m[4])
	if err != nil {
		err = NewFormatError("failed to decode hex data", err)
		return
	}

	// TODO: add extended id support, need an example log and a test.
	frame.Id = can.CanID{Id: uint32(id), Extended: false}
	frame.Data = rawData
	frame.Kind = can.CanDataFrame

	ts = time.Unix(unixSeconds, unixMicros*int64(time.Microsecond))

	return

}

// data is of the form
// 1698180835.318 0619D80564080EBE241
// the second part there is 3 nibbles (12 bits, 3 hex chars) for can ID,
// the rest is data.
// this regex does the processing. we precompile for speed.
var telemRegex = regexp.MustCompile(`^(\d+)\.(\d{3}) (\w{3})(\w+)$`)

func parseTelemLogLine(line string) (frame can.Frame, ts time.Time, err error) {
	frame = can.Frame{}
	ts = time.Unix(0, 0)
	// strip trailng newline since we rely on it being gone
	line = strings.TrimSpace(line)

	a := telemRegex.FindStringSubmatch(line)
	if a == nil || len(a) != 5 {
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
	ts = time.Unix(unixSeconds, unixMillis*int64(time.Millisecond))

	// VALIDATION STEP: sometimes the data gets really whack, but remains valid.
	// We check that the time is between 2017 and 2032.
	// Realistically we will not be using this software then.
	// TODO: add this

	id, err := strconv.ParseUint(a[3], 16, 16)
	if err != nil {
		err = NewFormatError("failed to parse id", err)
		return
	}

	if len(a[4])%2 != 0 {
		// odd hex chars, protect against a panic
		err = NewFormatError("wrong amount of hex chars", nil)
		return
	}
	rawData, err := hex.DecodeString(a[4])
	if err != nil {
		err = NewFormatError("failed to parse hex data", err)
		return
	}
	frame = can.Frame{
		Id:   can.CanID{Id: uint32(id), Extended: false},
		Data: rawData,
		Kind: can.CanDataFrame,
	}
	return frame, ts, nil

}

// BusEventParser is a function that takes a string and returns a busevent.
type BusEventParser func(string) (skylab.BusEvent, error)

// skylabify JSON parser.
func parseSkylabifyLogLine(input string) (skylab.BusEvent, error) {
	var b = skylab.BusEvent{}
	err := json.Unmarshal([]byte(input), &b)
	return b, err
}

// frameParseToBusEvent takes a line parser (that returns a can frame)
// and makes it return a busEvent instead.
func frameParseToBusEvent(fun CanFrameParser) BusEventParser {
	return func(s string) (skylab.BusEvent, error) {
		var b = skylab.BusEvent{}
		frame, ts, err := fun(s)
		if err != nil {
			return b, err
		}
		b.Timestamp = ts
		b.Data, err = skylab.FromCanFrame(frame)
		if err != nil {
			return b, err
		}
		b.Name = b.Data.String()
		return b, nil
	}
}

var ParsersMap = map[string]BusEventParser{
	"telem":   frameParseToBusEvent(parseTelemLogLine),
	"candump": frameParseToBusEvent(parseCanDumpLine),
	"json":    parseSkylabifyLogLine,
}
