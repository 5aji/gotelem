//go:build ignore
// +build ignore

// this file is a generator for skylab code.
package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// SkylabFile is a yaml file from skylab.
type SkylabFile struct {
	Packets []PacketDef
	Boards []BoardSpec

}

type BoardSpec struct {
	Name string
	Transmit []string
	Recieve []string
}

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
// constants for packet IDs or a map.


var test = `
packets:
  - name: dashboard_pedal_percentages
    description: ADC values from the brake and accelerator pedals.
    id: 0x290
    endian: little
    frequency: 10
    data:
      - name: reason1
        type: bitfield
        bits:
          - name: OVERVOLT
          - name: UNDERVOLT
          - name: OVERTEMP
          - name: TEMP_DISCONNECT
          - name: COMM_FAIL
      - name: reason2
        type: bitfield
        bits:
          - name: HARDWARE
          - name: KILL_PACKET
          - name: UKNOWN
          - name: OVERCURRENT
          - name: PRECHARGE_FAIL
          - name: AUX_OVER_UNDER
      - name: module
        type: uint16_t
      - name: value
        type: float
  - name: bms_module
    description: Voltage and temperature for a single module
    id: 0x01C
    endian: little
    repeat: 36
    offset: 1
    frequency: 2
    data:
      - name: voltage
        type: float
        units: V
        conversion: 1
      - name: temperature
        type: float
        units: C
        conversion: 1
`


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

func (d *DataField) ToStructMember(parentName string) string {
	
	if d.Type == "bitfield" {
		bfStructName := parentName + toCamelInitCase(d.Name, true)
		return toCamelInitCase(d.Name, true) + " " + bfStructName
	} else {
		return toCamelInitCase(d.Name, true) + " " + typeMap[d.Type]
	}
}

func (d *DataField) MakeMarshal(offset int) string {

	fieldName := toCamelInitCase(d.Name, true)
	if d.Type == "uint8_t" || d.Type == "int8_t" {
		return fmt.Sprintf("b[%d] = p.%s", offset, fieldName)
	} else if d.Type == "bitfield" {
		return fmt.Sprintf("b[%d] = p.%s.Marshal()", offset,fieldName)
	} else if d.Type == "float" {

		return fmt.Sprintf("float32ToBytes(b[%d:], p.%s, false)", offset, fieldName)

	} else if t ,ok := typeMap[d.Type]; ok {
		// it's uint or int of some kind, use endian to write it.
		return fmt.Sprintf("binary.LittleEndian.Put%s(b[%d:], p.%s)", toCamelInitCase(t, true), offset, fieldName)
	}
	return "panic(\"failed to do it\")\n"
}


func (d *DataField) MakeUnmarshal(offset int) string {

	fieldName := toCamelInitCase(d.Name, true)
	if d.Type == "uint8_t" || d.Type == "int8_t" {
		return fmt.Sprintf("p.%s = b[%d]", fieldName, offset)
	} else if d.Type == "bitfield" {
		return fmt.Sprintf("p.%s.Unmarshal(b[%d])", fieldName, offset)
	} else if d.Type == "float" {

		return fmt.Sprintf("p.%s = float32FromBytes(b[%d:], false)", fieldName, offset)

	} else if t ,ok := typeMap[d.Type]; ok {
		// it's uint or int of some kind, use endian to write it.
		// FIXME: support big endian
		return fmt.Sprintf("p.%s = binary.LittleEndian.%s(b[%d:])", fieldName, toCamelInitCase(t, true), offset)
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
{{ define "packet" }}
{{- $structName := camelCase .Name true}}

{{- /* generate any bitfield structs */ -}}
{{range .Data -}}
{{ if .Bits -}}
{{- $bfname := (printf "%s%s" $structName (camelCase .Name true)) }}
type {{$bfname}} struct {
	{{- range $el := .Bits}}
	{{camelCase $el.Name true}} bool
	{{- end}}
}

func (p *{{$bfname}}) Marshal() byte {
	var b byte
	{{- range $idx, $el := .Bits}}
	{{- $bitName := camelCase $el.Name true}}
	if p.{{$bitName}} {
		b |= 1 << {{$idx}}
	}
	{{- end}}
	return b
}

func (p *{{$bfname}}) Unmarshal(b byte) {
	{{- range $idx, $el := .Bits}}	
	{{- $bitName := camelCase $el.Name true }}
	p.{{$bitName}} = (b & (1 << {{ $idx }})) != 0
	{{- end}}
}
{{end}}
{{- end}}

// {{$structName}} is {{.Description}}
type {{$structName}} struct {
{{- range .Data}}
	{{ if .Units -}} // {{.Conversion}} {{.Units}} {{- end }}
	{{.ToStructMember $structName }}
{{- end}}
{{- if .Repeat }}
	// Idx is the packet index. The accepted range is 0-{{.Repeat}}
	Idx uint32
{{- end }}
}

func (p *{{$structName}}) Id() uint32 {
{{- if .Repeat }}
	return {{ printf "0x%X" .Id }} + p.Idx
{{- else }}
	return {{ printf "0x%X" .Id }}
{{- end }}
}

func (p *{{$structName}}) Size() uint {
	return {{.CalcSize}}
}

func (p *{{$structName}}) MarshalPacket() ([]byte, error) {
	b := make([]byte, {{ .CalcSize }})
{{.MakeMarshal}}
	return b, nil
}

func (p *{{$structName}}) UnmarshalPacket(b []byte) error {
{{.MakeUnmarshal}}
	return nil
}

func (p *{{$structName}}) String() string {
	return ""
}

{{ end }}

{{- /* begin actual file template */ -}}

// generated by gen_skylab.go at {{ Time }} DO NOT EDIT!

package skylab

import (
	"errors"
	"encoding/binary"
)

type SkylabId uint32

const (
{{- range .Packets }}
	{{camelCase .Name true}}Id SkylabId = {{.Id | printf "0x%X"}}
{{- end}}
)

// list of every packet ID. can be used for O(1) checks.
var idMap = map[uint32]bool{
	{{ range $p := .Packets -}}
	{{ if $p.Repeat }} 
	{{ range $idx := Nx (int $p.Id) $p.Repeat $p.Offset -}}
	{{ $idx | printf "0x%X"}}: true,
	{{ end }}
	{{- else }}
	{{ $p.Id | printf "0x%X" }}: true,
	{{- end}}
	{{- end}}
}

func FromCanFrame(id uint32, data []byte) (Packet, error) {
	if !idMap[id] {
		return nil, errors.New("Unknown Id")
	}
	switch id {
{{- range $p := .Packets }}
	{{- if $p.Repeat }}
	case {{ Nx (int $p.Id) $p.Repeat $p.Offset | mapf "0x%X" | strJoin ", " -}}:
		var res *{{camelCase $p.Name true}}
		res.UnmarshalPacket(data)
		res.Idx = id - uint32({{camelCase $p.Name true}}Id)
		return res, nil
	{{- else }}
	case {{ $p.Id | printf "0x%X" }}:
		var res *{{camelCase $p.Name true}}
		res.UnmarshalPacket(data)
		return res, nil
	{{- end}}
{{- end}}
	}

	return nil, errors.New("failed to match Id, something is really wrong!")
}
{{range .Packets -}}
{{template "packet" .}}
{{- end}}
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

// N takes a start and stop value and returns a stream of 
// [start, end), including the starting value but excluding the end value.
func N(start, end int) (stream chan int) {
    stream = make(chan int)
    go func() {
        for i := start; i < end; i++ {
            stream <- i
        }
        close(stream)
    }()
    return
}


// Nx takes a start, a quantity, and an offset and returns a stream
// of `times` values which count from start and increment by `offset` each
// time.
func Nx (start, times, offset int) (elems []int) {
	elems = make([]int, times)
	for i := 0; i < times; i++ {
		elems[i] = start + offset * i
	}
	return 
}

// dumb function for type conversion between uint32 to integer
// used for converting packet id to int for other functions internally.
func uint32ToInt(i uint32) (o int) {
	return int(i)
}


// strJoin is a remapping of strings.Join so that we can use
// it in a pipeline more easily.
func strJoin(delim string, elems []string) string {
	return strings.Join(elems, delim)
}

// mapf takes a slice of items and runs printf on each using the given format.
// it is basically mapping a slice of things to a slice of strings using a format string).
func mapf(format string, els []int) []string {
	resp := make([]string, len(els))
	for idx := range els {
		resp[idx] = fmt.Sprintf(format, els[idx])
	}
	return resp
}

func main() {
	// read path as the first arg, glob it for yamls, read each yaml into a skylabFile.
	// then take each skylab file, put all the packets into one big array.
	// then we need to make a header template.
	v := &SkylabFile{}

	err := yaml.Unmarshal([]byte(test), v)
	if err != nil {
		fmt.Printf("err %v", err)
	}

	fnMap := template.FuncMap{
		"camelCase": toCamelInitCase,
		"Time": time.Now,
		"N": N,
		"Nx": Nx,
		"int": uint32ToInt,
		"strJoin": strJoin,
		"mapf": mapf,
	}
	tmpl, err := template.New("skylab").Funcs(fnMap).Parse(templ)

	if err != nil {
		panic(err)
	}


	f, err := os.Create("skylab_gen.go")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(f, v)

	if err != nil {
		panic(err)
	}

}
