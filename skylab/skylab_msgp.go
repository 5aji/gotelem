package skylab

//go:generate msgp -unexported

// internal structure for handling
type msgpRawEvent struct {
	Timestamp float64 `msg:"ts"`
	Id        uint32  `msg:"id"`
	Data      []byte  `msg:"data"`
}

// internal structure to represent a raw can packet over the network.
// this is what's sent over the solar car to lead xbee connection
// for brevity while still having some robustness.
type msgpRawPacket struct {
	Id   uint32 `msg:"id"`
	Data []byte `msg:"data"`
}
