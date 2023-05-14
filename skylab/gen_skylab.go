//go:build ignore
// +build ignore

// this file is a generator for skylab code.
package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v3"
)

// data field.
type DataField struct {
	Name       string
	Type       string
	Units      string // mostly for documentation
	Conversion float32
	Bits       []struct {
		Name string
	}
}

// a PacketDef is a full can packet.
type PacketDef struct {
	Name        string
	Description string
	Id          uint32
	BigEndian   bool
	Repeat      int
	Offset      int
	Data        []DataField
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
		// the opposite of above.

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

var test = `
packets:
  - name: dashboard_pedal_percentages
    description: ADC values from the brake and accelerator pedals.
    id: 0x290
    endian: little
    frequency: 10
    data:
      - name: blah_blah
        type: uint32_t
      - name: accel_pedal_value
        type: float
      - name: brake_pedal_value
        type: uint8_t
`

type SkylabFile struct {
	Packets []PacketDef
}

var typeMap = map[string]string{
	"uint16_t": "uint16",
	"uint32_t": "uint32",
	"uint64_t": "uint64",
	"uint8_t":  "uint8",
	"float":    "float32",

	"int16_t": "int16",
	"int32_t": "int32",
	"int64_t": "int64",
	"int8_t":  "int8",
}

var typeSizeMap = map[string]uint{
	"uint16_t": 2,
	"uint32_t": 4,
	"uint64_t": 8,
	"uint8_t":  1,
	"float":    4,

	"int16_t":  2,
	"int32_t":  4,
	"int64_t":  8,
	"int8_t":   1,
	"bitfield": 1,
}

func (d *DataField) ToStructMember() string {
	if d.Type != "bitfield" {
		return toCamelInitCase(d.Name, true) + " " + typeMap[d.Type]
	}
	// it's a bitfield, things are more complicated.
	slog.Warn("bitfields are skipped for now")
	return ""
}

func (d *DataField) MakeMarshal(offset int) string {

	if d.Type == "uint8_t" || d.Type == "int8_t" {
		return fmt.Sprintf("b[%d] = p.%s", offset, toCamelInitCase(d.Name, true))
	} else if d.Type == "bitfield" {
		return "panic(\"bitfields don't work\")"
	} else if d.Type == "float" {

		return fmt.Sprintf("float32ToBytes(b[%d:], p.%s, false)", offset, toCamelInitCase(d.Name, true))

	} else if t ,ok := typeMap[d.Type]; ok {
		// it's uint or int of some kind, use endian to write it.
		return fmt.Sprintf("binary.LittleEndian.Put%s(b[%d:], p.%s)", toCamelInitCase(t, true), offset, toCamelInitCase(d.Name, true))
	}
	return "panic(\"failed to do it\")\n"
}


func (d *DataField) MakeUnmarshal(offset int) string {

		if d.Type == "uint8_t" || d.Type == "int8_t" {
			return fmt.Sprintf("p.%s = b[%d]", toCamelInitCase(d.Name, true), offset)
		} else if d.Type == "bitfield" {

		} else if d.Type == "float" {

			return fmt.Sprintf("p.%s = float32FromBytes(b[%d:], false)", toCamelInitCase(d.Name, true), offset)

		} else if t ,ok := typeMap[d.Type]; ok {
			// it's uint or int of some kind, use endian to write it.
			return fmt.Sprintf("p.%s = binary.LittleEndian.%s(b[%d:])", toCamelInitCase(d.Name, true), toCamelInitCase(t, true), offset)
		}
		panic("unhandled type")
}



func (p PacketDef) CalcSize() int {
	// makes a function that returns the size of the code.

	var size int = 0
	for _, val := range p.Data {
		size += int(typeSizeMap[val.Type])
	}

	return size
}


func (p PacketDef) MakeMarshal() string {
	var buf strings.Builder

	var offset int = 0
	// we have a b []byte as the correct-size byte array to store in.
	// and the packet itself is represented as `p`
	for _, val := range p.Data {

		buf.WriteRune('\t')
		buf.WriteString(val.MakeMarshal(offset))
		buf.WriteRune('\n')

		// shift our offset so that our next write is good.
		offset += int(typeSizeMap[val.Type])
	}


	return buf.String()
}

func (p PacketDef) MakeUnmarshal() string {
	var buf strings.Builder


	var offset int = 0
	for _, val := range p.Data {

		buf.WriteRune('\t')
		buf.WriteString(val.MakeUnmarshal(offset))
		buf.WriteRune('\n')
		offset += int(typeSizeMap[val.Type])
	}

	return buf.String()
}

var templ = `
// go code generated! don't touch!
{{ $structName := camelCase .Name true}}
// {{$structName}} is {{.Description}}
type {{$structName}} struct {
{{- range .Data}}
	{{.ToStructMember}}
{{- end}}
}

func (p *{{$structName}}) Id() uint32 {
	return {{printf "0x%X" .Id}}
}

func (p *{{$structName}}) Size() int {
	return {{.CalcSize}}
}

func (p *{{$structName}}) Marshal() []byte {
	b = make([]byte, {{ .Size }})
{{.MakeMarshal}}
	return b
}

func (p *{{$structName}}) Unmarshal(b []byte) {
	{{.MakeUnmarshal}}	
}
`


// stolen camelCaser code. initCase = true means CamelCase, false means camelCase
func toCamelInitCase(s string, initCase bool) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	n := strings.Builder{}
	n.Grow(len(s))
	capNext := initCase
	for i, v := range []byte(s) {
		vIsCap := v >= 'A' && v <= 'Z'
		vIsLow := v >= 'a' && v <= 'z'
		if capNext {
			if vIsLow {
				v += 'A'
				v -= 'a'
			}
		} else if i == 0 {
			if vIsCap {
				v += 'a'
				v -= 'A'
			}
		}
		if vIsCap || vIsLow {
			n.WriteByte(v)
			capNext = false
		} else if vIsNum := v >= '0' && v <= '9'; vIsNum {
			n.WriteByte(v)
			capNext = true
		} else {
			capNext = v == '_' || v == ' ' || v == '-' || v == '.'
		}
	}
	return n.String()
}


// stolen float32 to bytes code

func main() {
	v := &SkylabFile{}

	err := yaml.Unmarshal([]byte(test), v)
	if err != nil {
		fmt.Printf("err %v", err)
	}

	fmt.Printf("%#v\n", v.Packets)

	fnMap := template.FuncMap{
		"camelCase": toCamelInitCase,
	}
	tmpl, err := template.New("packet").Funcs(fnMap).Parse(templ)

	if err != nil {
		panic(err)
	}

	err = tmpl.Execute(os.Stdout, v.Packets[0])

	if err != nil {
		panic(err)
	}

}
