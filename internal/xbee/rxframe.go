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
	if data[0] != byte(RxPktType) {
		return nil, fmt.Errorf("incorrect frame type 0x%x", data[0])
	}

	rx := &RxFrame{
		Source:  binary.BigEndian.Uint64(data[1:]),
		Payload: data[12:],
	}
	// RX options
	opt := data[11]
	// todo: use this
	if (opt & 0x1) == 1 {
		rx.ACK = true
	}
	if (opt & 0x2) == 1 {
		rx.BCast = true
	}

	return rx, nil
}
