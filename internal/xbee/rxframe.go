package xbee

import (
	"encoding/binary"
	"fmt"
)

type RxFrame struct {
	Source  uint64
	ACK     bool
	BCast   bool
	Payload []byte
}

func ParseRxFrame(data []byte) (*RxFrame, error) {
	// data is the binary that *isn't* part of the frame
	// i.e it excludes start delimiter, length, and checksum.

	// check the frame type (data[0])
	if data[0] != byte(RxPkt) && data[0] != byte(RxPktExpl) {
		return nil, fmt.Errorf("incorrect frame type 0x%x", data[0])
	}

	rx := &RxFrame{
		Source:  binary.BigEndian.Uint64(data[1:]),
		Payload: data[12:],
	}
	// RX options
	opt := data[11]
	// todo: use this
	fmt.Print(opt)

	return rx, nil
}
