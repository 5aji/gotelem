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
		bits = binary.BigEndian.Uint32(b)
	} else {
		bits = binary.LittleEndian.Uint32(b)
	}
	return math.Float32frombits(bits)
}

// Packet is any Skylab-generated packet.
type Packet interface {
	MarshalPacket() ([]byte, error)
	UnmarshalPacket(p []byte) error
	CANId() (uint32, error)
	Size() uint
	String() string
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
func ToCanFrame(p Packet) (id uint32, data []byte, err error) {

	id, err = p.CANId()
	if err != nil {
		return
	}
	data, err = p.MarshalPacket()
	return
}

// ---- other wire encoding business ----

// internal structure for partially decoding json object.
type jsonRawEvent struct {
	Timestamp float64
	Id        uint32
	Name      string
	Data      json.RawMessage
}

// BusEvent is a timestamped Skylab packet
type BusEvent struct {
	Timestamp float64 `json:"ts"`
	Id        uint64  `json:"id"`
	Name      string  `json:"name"`
	Data      Packet  `json:"data"`
}

// FIXME: handle Name field.
func (e *BusEvent) MarshalJSON() (b []byte, err error) {
	// create the underlying raw event
	j := &jsonRawEvent{
		Timestamp: e.Timestamp,
		Id:        uint32(e.Id),
		Name:      e.Data.String(),
	}
	// now we use the magic Packet -> map[string]interface{} function
	j.Data, err = json.Marshal(e.Data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(j)

}

func (e *BusEvent) UnmarshalJSON(b []byte) error {
	var jRaw *jsonRawEvent

	err := json.Unmarshal(b, jRaw)

	if err != nil {
		return err
	}

	e.Timestamp = jRaw.Timestamp
	e.Id = uint64(jRaw.Id)
	e.Data, err = FromJson(jRaw.Id, jRaw.Data)
	e.Name = e.Data.String()

	return err
}

// FIXME: handle name field.
func (e *BusEvent) MarshalMsg(b []byte) ([]byte, error) {

	// we need to send the bytes as a []byte instead of
	// an object like the JSON one (lose self-documenting)
	data, err := e.Data.MarshalPacket()
	if err != nil {
		return nil, err
	}
	rawEv := &msgpRawEvent{
		Timestamp: e.Timestamp,
		Id:        uint32(e.Id),
		Data:      data,
	}

	return rawEv.MarshalMsg(b)
}

func (e *BusEvent) UnmarshalMsg(b []byte) ([]byte, error) {
	rawEv := &msgpRawEvent{}
	remain, err := rawEv.UnmarshalMsg(b)
	if err != nil {
		return remain, err
	}
	e.Timestamp = rawEv.Timestamp
	e.Id = uint64(rawEv.Id)
	e.Data, err = FromCanFrame(rawEv.Id, rawEv.Data)
	e.Name = e.Data.String()

	return remain, err
}

// we need to be able to parse the JSON as well.  this is done using the
// generator since we can use the switch/case thing since it's the fastest
