// Package xbee implements xbee API encoding and decoding.

// It encodes and decodes
// API frames from io.Writer and io.Reader by providing a WriteFrame function and
// a scanner.split function. It also includes internal packets for using the API.
// For end-users, it provides a simple net.Conn-like interface that can write
// and read arbitrary bytes (to be used by a higher level protocol)
package xbee

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

// the frames have an outer shell - we will make a function that takes
// an inner frame element and wraps it in the appropriate headers.

// first, we should make it take the frame directly, so we make an interface
// that represents "framable" things. note that bytes.Buffer also fulfils this.

// Frameable is an object that can be sent in an XBee Frame. An XBee Frame
// consists of a start delimiter, length, the payload, and a checksum.
type Frameable interface {
	// returns the API identifier for this frame.
	// encodes this frame correctly.
	Bytes() []byte
}

// calculateChecksum is a helper function to calculate the 1-byte checksum of a data range.
// the data range does not include the start delimiter, or the length uint16 (only the frame payload)
func calculateChecksum(data []byte) byte {
	var sum byte
	for _, v := range data {
		sum += v
	}
	return 0xFF - sum
}

// writeXBeeFrame takes some bytes and wraps it in an XBee frame.
//
// An XBee frame has a start delimiter, followed by the length of the payload,
// then the payload itself, and finally a checksum.
func writeXBeeFrame(w io.Writer, data []byte) (n int, err error) {

	frame := make([]byte, len(data)+4)
	frame[0] = 0x7E

	binary.BigEndian.PutUint16(frame[1:], uint16(len(data)))

	copy(frame[3:], data)

	chk := calculateChecksum(data)

	frame[len(frame)-1] = chk
	return w.Write(frame)
}

// now we can describe frames in other files that implement Frameable.
// the remaining challenge is reception and actual API frames.
// xbee uses the first byte of the "frame data" as the API identifier or command.

//go:generate stringer -output=api_frame_cmd.go -type xbeeCmd
type XBeeCmd byte

const (
	// commands sent to the xbee s3b

	ATCmdType        XBeeCmd = 0x08 // AT Command
	ATCmdQueueType   XBeeCmd = 0x09 // AT Command - Queue Parameter Value
	TxReqType        XBeeCmd = 0x10 // TX Request
	TxReqExplType    XBeeCmd = 0x11 // Explicit TX Request
	RemoteCmdReqType XBeeCmd = 0x17 // Remote Command Request
	// commands recieved from the xbee

	ATCmdResponseType XBeeCmd = 0x88 // AT Command Response
	ModemStatusType   XBeeCmd = 0x8A // Modem Status
	TxStatusType      XBeeCmd = 0x8B // Transmit Status
	RouteInfoType     XBeeCmd = 0x8D // Route information packet
	AddrUpdateType    XBeeCmd = 0x8E // Aggregate Addressing Update
	RxPktType         XBeeCmd = 0x90 // RX Indicator (AO=0)
	RxPktExplType     XBeeCmd = 0x91 // Explicit RX Indicator (AO=1)
	IOSampleType      XBeeCmd = 0x92 // Data Sample RX Indicator
	NodeIdType        XBeeCmd = 0x95 // Note Identification Indicator
	RemoteCmdRespType XBeeCmd = 0x97 // Remote Command Response
)

// AT commands are hard, so let's write out all the major ones here

// Now we will implement receiving packets from a character stream.
// we first need to make a thing that produces frames from a stream using a scanner.

// this is a split function for bufio.scanner. It makes it easier to handle the FSM
// for extracting data from a stream. For the Xbee, this means that we must
// find the magic start character, (check that it's escaped), read the length,
// and then ensure we have enough length to finish the token, requesting more data
// if we do not.
//
// see https://pkg.go.dev/bufio#SplitFunc for more info
// https://medium.com/golangspec/in-depth-introduction-to-bufio-scanner-in-golang-55483bb689b4
func xbeeFrameSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		// there's no data, request more.
		return 0, nil, nil
	}

	if startIdx := bytes.IndexByte(data, 0x7E); startIdx >= 0 {
		// we have a start character. see if we can get the length.
		// we add 4 since start delimiter (1) + length (2) + checksum (1).
		// the length inside the packet represents the frame data only.
		if len(data[startIdx:]) < 3 {
			// since we don't have enough bytes to get the length, instead we
			// will discard all data before the start index
			return startIdx, nil, nil
		}
		// FIXME: add bounds checking! this can panic.
		var frameLen = int(binary.BigEndian.Uint16(data[startIdx+1:startIdx+3])) + 4
		if len(data[startIdx:]) < frameLen {
			// we got the length, but there's not enough data for the frame. we can trim the
			// data that came before the start, but not return a token.
			return startIdx, nil, nil
		}
		// there is enough data to pull a frame.
		// todo: check checksum here? we can return an error.
		return startIdx + frameLen, data[startIdx : startIdx+frameLen], nil
	}
	// we didn't find a start character in our data, so request more. trash everythign given to us
	return len(data), nil, nil
}

// parseFrame takes a framed packet and returns the contents after checking the
// checksum and start delimiter.
func parseFrame(frame []byte) ([]byte, error) {
	if frame[0] != 0x7E {
		return nil, errors.New("incorrect start delimiter")
	}
	fsize := len(frame)
	if calculateChecksum(frame[3:fsize-1]) != frame[fsize] {
		return nil, errors.New("checksum mismatch")
	}
	return frame[3 : fsize-1], nil
}

// stackup
// low level readwriter (serial or IP socket)
// XBee library layer (frame encoding/decoding to/from structs)
// AT command layer (for setup/control)
// xbee conn-like layer (ReadWriter + custom control functions)
// application marshalling format (msgpack or json or gob)
// application
