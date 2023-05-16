package skylab

import (
	"encoding/binary"
	"math"
)

/*
This file provides helpers used for serializing and deserializing skylab packets.
It contains common code and interfaces.
*/


func float32ToBytes(b []byte, f float32, bigEndian bool) {
	bits := math.Float32bits(f)
	if bigEndian {
		binary.BigEndian.PutUint32(b, bits)
	} else {
		binary.LittleEndian.PutUint32(b, bits)
	}
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

// Packet is any Skylab-generated packet.
type Packet interface {
	MarshalPacket() ([]byte, error)
	UnmarshalPacket(p []byte) error
	Id() uint32
	Size() uint
	String() string
}

type Marshaler interface {
	MarshalPacket() ([]byte, error)
}
type Unmarshaler interface {
	UnmarshalPacket(p []byte) error
}

type Ider interface {
	Id() uint32
}

type Sizer interface {
	Size() uint
}

// CanSend takes a packet and makes a Can frame.
func CanSend(p Packet) error {

	return nil
}
