package can

type Frame struct {
	Id   uint32
	Data []byte
	Kind Kind
}

//go:generate stringer -output=frame_kind.go -type Kind
type Kind uint8

const (
	SFF Kind = iota // Standard ID Frame
	EFF             // Extended ID Frame
	RTR             // Remote Transmission Request Frame
	ERR             // Error Frame
)

type CanFilter struct {
	Id       uint32
	Mask     uint32
	Inverted bool
}

type CanSink interface {
	Send(*Frame) error
}

type CanSource interface {
	Recv() (*Frame, error)
}

type CanTransciever interface {
	CanSink
	CanSource
}
