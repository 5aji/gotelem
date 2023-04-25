package xbee

import (
	"bytes"
	"fmt"
	"encoding/binary"
)

// this file contains some AT command constants and
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
