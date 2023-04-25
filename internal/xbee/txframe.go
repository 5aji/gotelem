package xbee

import (
	"bytes"
	"encoding/binary"
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
