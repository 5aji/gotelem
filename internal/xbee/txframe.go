package xbee

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// transmissions to this address are instead broadcast
const BroadcastAddr = 0xFFFF

type TxFrame struct {
	Id          byte
	Destination uint64
	BCastRadius uint8
	Options     uint8
	Payload     []byte
}

func (txFrame *TxFrame) Bytes() []byte {
	buf := new(bytes.Buffer)

	buf.WriteByte(byte(TxReqType))

	buf.WriteByte(txFrame.Id)

	a := make([]byte, 8)
	binary.BigEndian.PutUint64(a, txFrame.Destination)
	buf.Write(a)

	// write the reserved part.
	buf.Write([]byte{0xFF, 0xFE})

	// write the radius
	buf.WriteByte(txFrame.BCastRadius)

	buf.WriteByte(txFrame.Options)

	buf.Write(txFrame.Payload)

	return buf.Bytes()
}

// we also handle transmit status response frames here.
// these are emitted by the xbee when the status of the tx packet is known.
// it has an Id that matches it to the corressponding transmit request.
type TxStatus uint8

//go:generate stringer -output=txStatus.go -type TxStatus
const (
	TxStatusSuccess  TxStatus = 0x00
	TxStatusNoACK    TxStatus = 0x01
	TxStatusCCAFail  TxStatus = 0x02
	TxStatusIndirect TxStatus = 0x03
	TxStatusTxFail   TxStatus = 0x04
	TxStatusACKFail  TxStatus = 0x21
	// TODO: finish this
)

type TxStatusFrame struct {
	Id     uint8    // the Frame identifier that this status frame represents.
	Status TxStatus // the status itself - TxStatus is a stringable.
}

func ParseTxStatusFrame(data []byte) (*TxStatusFrame, error) {
	if data[0] != byte(TxStatusType) {
		return nil, fmt.Errorf("incorrect frame type for Tx status frame 0x%x", data[0])
	}

	status := &TxStatusFrame{
		Id:     data[1],
		Status: TxStatus(data[2]),
	}

	return status, nil
}
