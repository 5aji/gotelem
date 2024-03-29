//go:build ignore
// +build ignore

// this file is a generator for skylab code.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v3"
)

// SkylabFile is a yaml file from skylab.
type SkylabFile struct {
	Packets []PacketDef `yaml:"packets,omitempty" json:"packets,omitempty"`
	Boards  []BoardDef  `yaml:"boards,omitempty" json:"boards,omitempty"`
}

type BoardDef struct {
	Name     string   `yaml:"name,omitempty" json:"name,omitempty"`
	Transmit []string `yaml:"transmit,omitempty" json:"transmit,omitempty"`
	Receive  []string `yaml:"receive,omitempty" json:"receive,omitempty"`
}

// data field.
type FieldDef struct {
	Name       string  `yaml:"name,omitempty" json:"name,omitempty"`
	Type       string  `yaml:"type,omitempty" json:"type,omitempty"`
	Units      string  `yaml:"units,omitempty" json:"units,omitempty"`
	Conversion float32 `yaml:"conversion,omitempty" json:"conversion,omitempty"`
	Bits       []struct {
		Name string `yaml:"name,omitempty" json:"name,omitempty"`
	} `yaml:"bits,omitempty" json:"bits,omitempty"`
}

// a PacketDef is a full can packet.
type PacketDef struct {
	Name        string     `yaml:"name,omitempty" json:"name,omitempty"`
	Description string     `yaml:"description,omitempty" json:"description,omitempty"`
	Id          uint32     `yaml:"id,omitempty" json:"id,omitempty"`
	Endian      string     `yaml:"endian,omitempty" json:"endian,omitempty"`
	IsExtended  bool       `yaml:"is_extended,omitempty" json:"is_extended,omitempty"`
	Repeat      int        `yaml:"repeat,omitempty" json:"repeat,omitempty"`
	Offset      int        `yaml:"offset,omitempty" json:"offset,omitempty"`
	Data        []FieldDef `yaml:"data,omitempty" json:"data,omitempty"`
}

// we need to generate bitfield types.
// constants for packet IDs or a map.

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

func MapType(ctype string) string {
	return typeMap[ctype]
}

func (d *FieldDef) ToStructMember(parentName string) string {

	if d.Type == "bitfield" {
		bfStructName := parentName + toCamelInitCase(d.Name, true)
		return toCamelInitCase(d.Name, true) + " " + bfStructName
	} else {
		return toCamelInitCase(d.Name, true) + " " + typeMap[d.Type]
	}
}

func (d *FieldDef) MakeMarshal(offset int) string {

	fieldName := toCamelInitCase(d.Name, true)
	if d.Type == "uint8_t" || d.Type == "int8_t" {
		return fmt.Sprintf("b[%d] = p.%s", offset, fieldName)
	} else if d.Type == "bitfield" {
		return fmt.Sprintf("b[%d] = p.%s.MarshalByte()", offset, fieldName)
	} else if d.Type == "float" {

		return fmt.Sprintf("float32ToBytes(b[%d:], p.%s, false)", offset, fieldName)

	} else if t, ok := typeMap[d.Type]; ok {
		// it's uint or int of some kind, use endian to write it.
		if strings.HasPrefix(t, "i") {
			// this means it's a signed integer.
			// encoding/binary does not support putting signed ints, instead
			// we should cast it to unsigned and then use the unsigned int functions.
			return fmt.Sprintf("binary.LittleEndian.PutU%s(b[%d:], u%s(p.%s))", t, offset, t, fieldName)
		}
		return fmt.Sprintf("binary.LittleEndian.Put%s(b[%d:], p.%s)", toCamelInitCase(t, true), offset, fieldName)
	}
	return "panic(\"failed to do it\")\n"
}

func (d *FieldDef) MakeUnmarshal(offset int) string {

	fieldName := toCamelInitCase(d.Name, true)
	if d.Type == "uint8_t" || d.Type == "int8_t" {
		return fmt.Sprintf("p.%s = b[%d]", fieldName, offset)
	} else if d.Type == "bitfield" {
		return fmt.Sprintf("p.%s.UnmarshalByte(b[%d])", fieldName, offset)
	} else if d.Type == "float" {

		return fmt.Sprintf("p.%s = float32FromBytes(b[%d:], false)", fieldName, offset)

	} else if t, ok := typeMap[d.Type]; ok {
		// it's uint or int of some kind, use endian to read it.
		// FIXME: support big endian
		if strings.HasPrefix(t, "i") {
			// this means it's a signed integer.
			// encoding/binary does not support putting signed ints, instead
			// we should cast it to unsigned and then use the unsigned int functions.
			return fmt.Sprintf("p.%s = %s(binary.LittleEndian.U%s(b[%d:]))", fieldName, t, t, offset)
		}
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
func Nx(start, times, offset int) (elems []int) {
	elems = make([]int, times)
	for i := 0; i < times; i++ {
		elems[i] = start + offset*i
	}
	return
}

// dumb function for type conversion between uint32 to integer
// used for converting packet id to int for other functions internally.
func uint32ToInt(i uint32) (o int) {
	return int(i)
}

// strJoin is a remapping of strings.Join so that we can use
// it in a pipeline.
//
//	{{.Names | strJoin ", " }}
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

func idToString(p PacketDef) string {
	if p.Repeat > 0 {
		resp := make([]string, p.Repeat)
		for idx := 0; idx < p.Repeat; idx++ {
			resp[idx] = fmt.Sprintf("can.CanID{ Id: 0x%X, Extended: %t }", int(p.Id)+idx*p.Offset, p.IsExtended)
		}

		return strings.Join(resp, ",")

	} else {
		return fmt.Sprintf("can.CanID{ Id: 0x%X, Extended: %t }", p.Id, p.IsExtended)
	}
}

func main() {
	// read path as the first arg, glob it for yamls, read each yaml into a skylabFile.
	// then take each skylab file, put all the packets into one big array.
	// then we need to make a header template.
	v := &SkylabFile{}

	basePath, err := filepath.Abs(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Printf("skylab packet definition path is %s\n", basePath)

	fGlob := filepath.Join(basePath, "*.y?ml")
	files, err := filepath.Glob(fGlob)
	if err != nil {
		panic(err)
	}
	fmt.Printf("found %d files\n", len(files))
	for _, f := range files {
		fd, err := os.Open(f)
		if err != nil {
			fmt.Printf("failed to open file %s:%v\n", filepath.Base(f), err)
		}
		dec := yaml.NewDecoder(fd)
		newFile := &SkylabFile{}
		err = dec.Decode(newFile)
		if err != nil {
			panic(err)
		}
		fmt.Printf("%s: adding %d packets and %d boards\n", filepath.Base(f), len(newFile.Packets), len(newFile.Boards))
		v.Packets = append(v.Packets, newFile.Packets...)
		v.Boards = append(v.Boards, newFile.Boards...)
	}

	// we add any functions mapping we need here.
	fnMap := template.FuncMap{
		"camelCase":  toCamelInitCase,
		"Time":       time.Now,
		"N":          N,
		"Nx":         Nx,
		"int":        uint32ToInt,
		"strJoin":    strJoin,
		"mapf":       mapf,
		"maptype":    MapType,
		"json":       json.Marshal,
		"idToString": idToString,
	}

	tmpl, err := template.New("golang.go.tmpl").Funcs(fnMap).ParseGlob("templates/*.go.tmpl")

	if err != nil {
		panic(err)
	}

	f, err := os.Create("skylab_gen.go")
	defer f.Close()
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(f, v)

	if err != nil {
		panic(err)
	}

	tests := tmpl.Lookup("golang_tests.go.tmpl")
	if tests == nil {
		panic("tests not found")
	}

	testF, err := os.Create("skylab_gen_test.go")
	defer f.Close()
	if err != nil {
		panic(err)
	}
	tests.Execute(testF, v)

}
