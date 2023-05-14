package skylab

import (
	"math"
	"encoding/binary"
)


func float32ToBytes(b []byte, f float32, bigEndian bool) {
	bits := math.Float32bits(f)
	if bigEndian {
		binary.BigEndian.PutUint32(b, bits)
	} else {
		binary.LittleEndian.PutUint32(b, bits)
	}
	return
}

func float32FromBytes(b []byte, bigEndian bool) (f float32) {
	var bits uint32
	if bigEndian {
		binary.BigEndian.Uint32(b)
	} else {
		binary.LittleEndian.Uint32(b)
	}
	return math.Float32frombits(bits)
}


type Packet interface {
	MarshalPacket() ([]byte, error)
	UnmarshalPacket(p []byte) error
	Id() uint32
	Size() int
	String()
}
