// Package skylab provides CAN packet encoding and decoding information based off
// of skylab.yaml. It can convert packets to/from CAN raw bytes and JSON objects.
package skylab

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"time"

	// this is needed so that we can run make_skylab.go
	// without this, the yaml library will be removed
	// when we run `go mod tidy`
	"github.com/kschamplin/gotelem/internal/can"
	_ "gopkg.in/yaml.v3"
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
	Marshaler
	Unmarshaler
	Ider
	Sizer
	fmt.Stringer // to get the name
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
	CanId() (can.CanID, error)
}

// Sizer allows for fast allocation.
type Sizer interface {
	Size() uint
}

// CanSend takes a packet and makes CAN framing data.
func ToCanFrame(p Packet) (f can.Frame, err error) {

	f.Id, err = p.CanId()
	if err != nil {
		return
	}
	f.Data, err = p.MarshalPacket()
	f.Kind = can.CanDataFrame
	return
}

// ---- other wire encoding business ----

// internal structure for partially decoding json object.
type RawJsonEvent struct {
	Timestamp int64           `json:"ts" db:"ts"`
	Name      string          `json:"name"`
	Data      json.RawMessage `json:"data"`
}

// BusEvent is a timestamped Skylab packet - it contains
type BusEvent struct {
	Timestamp time.Time `json:"ts"`
	Name      string    `json:"id"`
	Data      Packet    `json:"data"`
}

func (e BusEvent) MarshalJSON() (b []byte, err error) {
	// create the underlying raw event
	j := &RawJsonEvent{
		Timestamp: e.Timestamp.UnixMilli(),
		Name:      e.Name,
	}
	// now we use the magic Packet -> map[string]interface{} function
	// FIXME: this uses reflection and isn't good for the economy
	j.Data, err = json.Marshal(e.Data)
	if err != nil {
		return nil, err
	}

	return json.Marshal(j)

}

func (e *BusEvent) UnmarshalJSON(b []byte) error {
	j := &RawJsonEvent{}

	err := json.Unmarshal(b, j)

	if err != nil {
		return err
	}

	e.Timestamp = time.UnixMilli(j.Timestamp)
	e.Name = j.Name
	e.Data, err = FromJson(j.Name, j.Data)

	return err
}

// we need to be able to parse the JSON as well.  this is done using the
// generator since we can use the switch/case thing since it's the fastest

type UnknownIdError struct {
	id uint32
}

func (e *UnknownIdError) Error() string {
	return fmt.Sprintf("unknown id: %x", e.id)
}

type BadLengthError struct {
	expected uint32
	actual   uint32
}

func (e *BadLengthError) Error() string {
	return fmt.Sprintf("bad data length, expected %d, got %d", e.expected, e.actual)
}
