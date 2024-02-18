package logparsers

import (
	"encoding/hex"
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
	return fmt.Sprintf("%s", e.msg)

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
	b = skylab.BusEvent{}
	// dumpline looks like this:
	// (1684538768.521889) can0 200#8D643546
	// remove trailing newline
	dumpLine = strings.TrimSpace(dumpLine)
	segments := strings.Split(dumpLine, " ")

	var unixSeconds, unixMicros int64
	fmt.Sscanf(segments[0], "(%d.%d)", &unixSeconds, &unixMicros)

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

	b.Timestamp = time.Unix(unixSeconds, unixMicros)

	b.Data, err = skylab.FromCanFrame(frame)

	if err != nil {
		err = NewFormatError("failed to parse can frame", err)
		return
	}

	// set the name
	b.Name = b.Data.String()

	return
}

var telemRegex = regexp.MustCompile(`^(\d+).(\d{3}) (\w{3})(\w+)$`)
func parseTelemLogLine(line string) (b skylab.BusEvent, err error) {
	b = skylab.BusEvent{}
	// strip trailng newline since we rely on it being gone
	line = strings.TrimSpace(line)
	// data is of the form
	// 1698180835.318 0619D80564080EBE241
	// the second part there is 3 nibbles (12 bits, 3 hex chars) for can ID,
	// the rest is data.
	// this regex does the processing.

	// these files tend to get corrupted. there are all kinds of nasties that can happen.
	// defense against random panics
	defer func() {
		if r := recover(); r != nil {
			err = NewFormatError("caught panic", nil)
		}
	}()
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
	ts := time.Unix(unixSeconds, unixMillis*1e6)

	// VALIDATION STEP: sometimes the data gets really whack.
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

var ParsersMap = map[string]ParserFunc{
	"telem":   parseTelemLogLine,
	"candump": parseCanDumpLine,
}
