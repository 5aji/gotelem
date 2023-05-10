//go:build ignore

// this file is a generator for skylab code.
package main


import (
	"gopkg.in/yaml.v3"
)

type Field interface {
	Name() string

	Size() int // the size of the data.
	
	// returns something like
	// 	AuxVoltage uint16
	// used inside the packet struct
	Embed() string
	
	// returns 
	Marshal() string
	Decode() string
}

// this is a standard field, not a bitfield.
type DataField struct {
	Name string
	Type string
	Units string // mostly for documentation
	Conversion float32 
}


// a PacketDef is a full can packet.
type PacketDef struct {
	Name string
	Description string
	Id uint32
	BigEndian bool
	data: []Field
}

// we need to generate bitfield types.
// packet structs per each packet
// constancts for packet IDs or a map.


/*


example for a simple packet type
it also needs a json marshalling.

	type BMSMeasurement struct {
		BatteryVoltage uint16
		AuxVoltage uint16
		Current float32
	}

	func (b *BMSMeasurement)MarshalPacket() ([]byte, error) {
		pkt := make([]byte, b.Size())
		binary.LittleEndian.PutUint16(pkt[0:], b.BatteryVoltage * 0.01)
		binary.LittleEndian.PutUint16(pkt[2:],b.AuxVoltage * 0.001)
		binary.LittleEndian.PutFloat32(b.Current) // TODO: make float function
	}

	func (b *BMSMeasurement)UnmarshalPacket(p []byte) error {

	}

	func (b *BMSMeasurement) Id() uint32 {
		return 0x010
	}

	func (b *BMSMeasurement) Size() int {
		return 8
	}

	func (b *BMSMeasurement) String() string {
		return "blah blah"
	}

we also need some kind of mechanism to lookup data type.

	func getPkt (id uint32, data []byte) (Packet, error) {

		// insert really massive switch case statement here.
	}

*/
