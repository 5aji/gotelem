package skylab

import (
	"encoding/binary"
	"encoding/json"
	"math"
)

/*
This file provides helpers used for serializing and deserializing skylab packets.
It contains common code and interfaces.
*/

// float32ToBytes is an internal function used to encode a float value to bytes
func float32ToBytes(b []byte, f float32, bigEndian bool) {
	bits := math.Float32bits(f)
	if bigEndian {
		binary.BigEndian.PutUint32(b, bits)
	} else {
		binary.LittleEndian.PutUint32(b, bits)
	}
}

// float32FromBytes is an internal function used to decode float value from bytes
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
	CANId() (uint32, error)
	Size() uint
}

// Marshaler is a packet that can be marshalled into bytes.
type Marshaler interface {
	MarshalPacket() ([]byte, error)
}

// Unmarshaler is a packet that can be unmarshalled from bytes.
type Unmarshaler interface {
	UnmarshalPacket(p []byte) error
}

// Ider is a packet that can get its ID, based on the index of the packet, if any.
type Ider interface {
	CANId() (uint32, error)
}

// Sizer allows for fast allocation.
type Sizer interface {
	Size() uint
}

// CanSend takes a packet and makes CAN framing data.
func CanSend(p Packet) (id uint32, data []byte, err error) {

	id, err = p.CANId()
	if err != nil {
		return
	}
	data, err = p.MarshalPacket()
	return
}

// ---- JSON encoding business ----

type JSONPacket struct {
	Id uint32
	Data json.RawMessage
}

func ToJson(p Packet) (*JSONPacket, error) {

	d, err := json.Marshal(p)

	if err != nil {
		return nil, err
	}

	id, err := p.CANId()
	if err != nil {
		return nil, err
	}

	jp := &JSONPacket{Id: id, Data: d}

	return jp, nil
}

// we need to be able to parse the JSON as well.
// this is done using the generator since we can use the switch/case thing
// since it's the fastest
