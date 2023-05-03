package xbee

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// This code is handled slightly differently. Rather than use structs to represent
// the frame data, we instead use an interface that represents an AT command, and
// standalone functions that convert the AT command to frame data.
// this way, we can add command support without needing to compose structs.
// The downside is that it's more difficult to deal with.
//
// AT command responses are handled with a struct to get the response status code.
//

// an ATCmd is anything that has a Payload function that returns the given payload
// pre-formatted to be send over the wire, and a Cmd command that returns the 2-character
// ATcommand itself. It must also have a parse response function that takes the response struct.
type ATCmd interface {
	Payload() []byte            // converts the AT command options to the binary argument format
	Cmd() [2]rune               // returns the AT command string.
	Parse(*ATCmdResponse) error // takes a command response and parses the data into itself.
}

// The AT command, in its raw format. we don't handle parameters since it's dealt with by
// the interface.
type RawATCmd []byte

// implements frameable. this is kinda stupid but makes it more safe.
func (b RawATCmd) Bytes() []byte {
	return b
}

// EncodeATCommand takes an AT command and encodes it in the payload format.
// it takes the frame index (which can be zero) as well as if it should be queued or
// not. It encodes the AT command to be framed and sent over the wire and returns the packet
func encodeATCommand(cmd [2]rune, p []byte, idx uint8, queued bool) RawATCmd {
	// we encode a new byte slice that contains the cmd + payload concatenated correclty.
	// this is then used to make the command frame, which contains ID/Type/Queued or not.
	// the ATCmdFrame can be converted to bytes to be sent over the wire once framed.

	buf := new(bytes.Buffer)

	// we change the frame type based on if it's queued.
	if queued {
		buf.WriteByte(byte(ATCmdQueueType))
	} else {
		buf.WriteByte(byte(ATCmdType))
	}

	// next is the provided frame identifier, used for callbacks.
	buf.WriteByte(idx)

	buf.WriteByte(byte(cmd[0]))
	buf.WriteByte(byte(cmd[1]))
	// the payload can be empty. This would make it a query.
	buf.Write(p)

	return buf.Bytes()
}

type ATCmdResponse struct {
	Cmd    string
	Status ATCmdStatus
	Data   []byte
}

func ParseATCmdResponse(p []byte) (*ATCmdResponse, error) {

	if p[0] != 0x88 {
		return nil, fmt.Errorf("invalid frame type 0x%x", p[0])
	}
	resp := &ATCmdResponse{
		Cmd:    string(p[2:4]),
		Status: ATCmdStatus(p[4]),
		// TODO: check if this overflows when there's no payload.
		Data: p[5:],
	}

	return resp, nil
}

//go:generate stringer -output=at_cmd_status.go -type ATCmdStatus
type ATCmdStatus uint8

const (
	ATCmdStatusOK  ATCmdStatus = 0
	ATCmdStatusErr ATCmdStatus = 1

	ATCmdStatusInvalidCmd   ATCmdStatus = 2
	ATCmdStatusInvalidParam ATCmdStatus = 3
)

type RemoteATCmdReq struct {
	Destination uint64
	Options     uint8
}

func encodeRemoteATCommand(at ATCmd, idx uint8, queued bool, destination uint64) RawATCmd {

	// sizing take from
	buf := new(bytes.Buffer)

	buf.WriteByte(byte(RemoteCmdReqType))

	buf.WriteByte(idx)

	binary.Write(buf, binary.BigEndian, destination)

	binary.Write(buf, binary.BigEndian, uint16(0xFFFE))

	// set remote command options. if idx = 0, we set bit zero (disable ack)
	// if queued is true, we clear bit one (if false we set it)

	var options uint8 = 0
	if idx == 0 {
		options = options | 0x1
	}
	if !queued {
		options = options | 0x2
	}
	buf.WriteByte(options)

	// write AT command
	cmd := at.Cmd()
	buf.WriteByte(byte(cmd[0]))
	buf.WriteByte(byte(cmd[1]))

	// write payload.
	buf.Write(at.Payload())

	return buf.Bytes()
}

// let's actually define some AT commands now.

// TODO: should we just use a function.
