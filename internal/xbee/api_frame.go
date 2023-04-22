// Package xbee implements xbee API encoding and decoding.

// It encodes and decodes
// API frames from io.Writer and io.Reader by providing a WriteFrame function and
// a scanner.split function. It also includes
package xbee

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// the frames have an outer shell - we will make a function that takes
// an inner frame element and wraps it in the appropriate headers.

// first, we should make it take the frame directly, so we make an interface
// that represents "framable" things. note that bytes.Buffer also fulfils this.

type Frameable interface {
	// returns the API identifier for this frame.
	GetId() byte
	// encodes this frame correctly.
	Bytes() ([]byte, error)
}

// now we can describe our function that takes a framable and contains it + calculates checksums.
func calculateChecksum(data []byte) byte {
	var sum byte
	for _, v := range data {
		sum += v
	}
	return 0xFF - sum
}

func WriteFrame(w io.Writer, cmd Frameable) (n int, err error) {
	frame_data, err := cmd.Bytes()

	if err != nil {
		return
	}
	frame := make([]byte, len(frame_data)+4)
	frame[0] = 0x7E

	binary.BigEndian.PutUint16(frame[1:], uint16(len(frame_data)))

	copy(frame[3:], frame_data)

	chk := calculateChecksum(frame_data)

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

	ATCmd        XBeeCmd = 0x08 // AT Command
	ATCmdQueue   XBeeCmd = 0x09 // AT Command - Queue Parameter Value
	TxReq        XBeeCmd = 0x10 // TX Request
	TxReqExpl    XBeeCmd = 0x11 // Explicit TX Request
	RemoteCmdReq XBeeCmd = 0x17 // Remote Command Request
	// commands recieved from the xbee

	ATCmdResponse XBeeCmd = 0x88 // AT Command Response
	ModemStatus   XBeeCmd = 0x8A // Modem Status
	TxStatus      XBeeCmd = 0x8B // Transmit Status
	RouteInfoPkt  XBeeCmd = 0x8D // Route information packet
	AddrUpdate    XBeeCmd = 0x8E // Aggregate Addressing Update
	RxPkt         XBeeCmd = 0x90 // RX Indicator (AO=0)
	RxPktExpl     XBeeCmd = 0x91 // Explicit RX Indicator (AO=1)
	IOSample      XBeeCmd = 0x92 // Data Sample RX Indicator
	NodeId        XBeeCmd = 0x95 // Note Identification Indicator
	RemoteCmdResp XBeeCmd = 0x97 // Remote Command Response
)

// AT commands are hard, so let's write out all the major ones here

type ATCmdFrame struct {
	Id     byte
	Cmd    string
	Param  []byte
	Queued bool
}

// implement the frame stuff for us.
func (atFrame *ATCmdFrame) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)

	if atFrame.Queued {
		// queued (batched) at comamnds have different Frame type
		buf.WriteByte(byte(ATCmdQueue))

	} else {
		// normal frame type
		buf.WriteByte(byte(ATCmd))

	}

	buf.WriteByte(atFrame.Id)

	// write cmd, if it's the right length.
	if cmdLen := len(atFrame.Cmd); cmdLen != 2 {
		return nil, fmt.Errorf("AT command incorrect length: %d", cmdLen)
	}
	buf.Write([]byte(atFrame.Cmd))

	// write param.
	buf.Write(atFrame.Param)
	return buf.Bytes(), nil
}

// transmissions to this address are instead broadcast
const BroadcastAddr = 0xFFFF

type TxFrame struct {
	Id          byte
	Destination uint64
	BCastRadius uint8
	Options     uint8
	Payload     []byte
}

func (txFrame *TxFrame) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)

	buf.WriteByte(byte(TxReq))

	buf.WriteByte(txFrame.Id)

	a := make([]byte, 8)
	binary.LittleEndian.PutUint64(a, txFrame.Destination)
	buf.Write(a)

	// write the reserved part.
	buf.Write([]byte{0xFF, 0xFE})

	// write the radius
	buf.WriteByte(txFrame.BCastRadius)

	buf.WriteByte(txFrame.Options)

	buf.Write(txFrame.Payload)

	return buf.Bytes(), nil
}

type RemoteATCmdReq struct {
	ATCmdFrame
	Destination uint64
	Options     uint8
}

func (remoteAT *RemoteATCmdReq) Bytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(RemoteCmdReq))

	buf.WriteByte(remoteAT.Id)

	a := make([]byte, 8)
	binary.LittleEndian.PutUint64(a, remoteAT.Destination)
	buf.Write(a)

	// write the reserved part.
	buf.Write([]byte{0xFF, 0xFE})
	// write options
	buf.WriteByte(remoteAT.Options)

	// now, write the AT command and the data.
	buf.Write([]byte(remoteAT.Cmd))

	buf.Write(remoteAT.Param)

	return buf.Bytes(), nil

}

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
		// we have a start character. get the length.
		// we add 4 since start delimiter (1) + length (2) + checksum (1).
		// the length inside the packet represents the frame data only.
		var frameLen = binary.BigEndian.Uint16(data[startIdx+1:startIdx+3]) + 4
		if len(data[startIdx:]) < int(frameLen) {
			// we got the length, but there's not enough data for the frame. we can trim the
			// data that came before the start, but not return a token.
			return startIdx, nil, nil
		}
		// there is enough data to pull a frame.
		// todo: check checksum here? we can return an error.
		return startIdx + int(frameLen), data[startIdx : startIdx+int(frameLen)], nil
	}
	// we didn't find a start character in our data, so request more. trash everythign given to us
	return len(data), nil, nil
}
