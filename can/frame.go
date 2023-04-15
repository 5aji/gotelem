package can

type Frame struct {
	ID   uint32
	Data []uint8
	Kind Kind
}

//go:generate stringer -output=frame_kind.go -type Kind
type Kind uint8

const (
	SFF Kind = iota // Standard Frame Format
	EFF             // Extended Frame
	RTR             // remote transmission requests
	ERR             // Error frame.
)

// for routing flexibility

type CanSink interface {
	Send(Frame) error
}

type CanSource interface {
	Recv(Frame) error
}
