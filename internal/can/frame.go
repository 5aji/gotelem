// Package can provides generic CAN interfaces and types.
//
// It has a generic can Frame (packet), as well as a filter type.
// we also define standard interfaces for objects that can accept
// can frames. We can use this pattern to easily extend the capabilities of the program
// by writing "adapters" to various devices/formats (xbee, socketcan)
package can


type CanID struct {
	Id uint32
	Extended bool // since the id itself is not enough.
}
// Frame represents a protocol-agnostic CAN frame. The Id can be standard or extended,
// but if it is extended, the Kind should be EFF.
type Frame struct {
	Id CanID
	Data []byte
	Kind Kind
}


// TODO: should this be replaced
type CANFrame interface {
	Id() 
	Data() []byte
	Type() Kind
}

//go:generate stringer -output=frame_kind.go -type Kind

// Kind is the type of the can Frame
type Kind uint8

const (
	CanDataFrame Kind = iota // Standard ID Frame
	CanRTRFrame             // Remote Transmission Request Frame
	CanErrFrame             // Error Frame
)

// CanFilter is a basic filter for masking out data. It has an Inverted flag
// which indicates opposite behavior (reject all packets that match Id and Mask).
// The filter matches when (packet.Id & filter.Mask) == filter.Id
// TODO: is this needed anymore since we are using firmware based version instead?
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
