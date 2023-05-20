// Package can provides generic CAN interfaces and types.
//
// It has a generic can Frame (packet), as well as a filter type.
// we also define standard interfaces for objects that can accept
// can frames. We can use this pattern to easily extend the capabiltiies of the program
// by writing "adapters" to various devices/formats (xbee, sqlite, network socket, socketcan)
package gotelem

import (
	"fmt"
	"os"
	"time"
)

// Frame represents a protocol-agnostic CAN frame. The Id can be standard or extended,
// but if it is extended, the Kind should be EFF.
type Frame struct {
	Id   uint32
	Data []byte
	Kind Kind
}

//go:generate msgp
type CANFrame interface {
	Id() uint32
	Data() []byte
	Type() Kind
}

//go:generate stringer -output=frame_kind.go -type Kind

// Kind is the type of the can Frame
type Kind uint8

const (
	CanSFFFrame Kind = iota // Standard ID Frame
	CanEFFFrame             // Extended ID Frame
	CanRTRFrame             // Remote Transmission Request Frame
	CanErrFrame             // Error Frame
)

// CanFilter is a basic filter for masking out data. It has an Inverted flag
// which indicates opposite behavior (reject all packets that match Id and Mask).
// The filter matches when (packet.Id & filter.Mask) == filter.Id
type CanFilter struct {
	Id       uint32
	Mask     uint32
	Inverted bool
}

// CanSink is an object that can accept Frames to transmit.
type CanSink interface {
	Send(*Frame) error
}

// CanSource is an object that can receive Frames.
type CanSource interface {
	Recv() (*Frame, error)
}

// CanTransciever is an object that can both send and receive Frames.
type CanTransciever interface {
	CanSink
	CanSource
}

// CanWriter
type CanWriter struct {
	output *os.File
}

// send writes the frame to the file.
func (cw *CanWriter) Send(f *Frame) error {
	ts := time.Now().Unix()

	_, err := fmt.Fprintf(cw.output, "%d %X %X\n", ts, f.Id, f.Data)
	return err
}

func (cw *CanWriter) Close() error {
	return cw.output.Close()
}

func OpenCanWriter(name string) (*CanWriter, error) {
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	cw := &CanWriter{
		output: f,
	}
	return cw, nil
}
