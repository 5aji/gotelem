package xbee

import "encoding/binary"

// the frames have an outer shell - we will make a function that takes
// an inner frame element and wraps it in the appropriate headers.

// first, we should make it take the frame directly, so we make an interface
// that represents "framable" things. note that bytes.Buffer also fulfils this.

type Frameable interface {
	Bytes() []byte
}

// now we can describe our function that takes a framable and contains it + calculates checksums.
func calculateChecksum(data []byte) byte {
	var sum byte
	for _, v := range data {
		sum += v
	}
	return 0xFF - sum
}
func makeXbeeApiFrame(cmd Frameable) ([]byte, error) {
	dataBuf := cmd.Bytes()
	frameBuf := make([]byte, len(dataBuf)+4)

	// move data and construct the frame

	frameBuf[0] = 0x7E // start delimiter

	// length
	// todo: check endiannes (0x7e, msb lsb)
	binary.LittleEndian.PutUint16(frameBuf[1:3], uint16(len(dataBuf)))

	copy(frameBuf[3:], dataBuf)

	chksum := calculateChecksum(dataBuf)

	frameBuf[len(frameBuf)-1] = chksum

	return frameBuf, nil
}

// now we can describe frames in other files that implement Frameable. this makes trasmission complete.
// the remaining challenge is reception and actual API frames.
// xbee uses the first byte of the "frame data" as the API identifier or command.

//go:generate stringer -output=api_frame_cmd.go -type xbeeCmd
type xbeeCmd byte

const (
	// commands sent to the xbee s3b

	ATCmd          xbeeCmd = 0x08 // AT Command
	ATCmdQueuePVal xbeeCmd = 0x09 // AT Command - Queue Parameter Value
	TxReq          xbeeCmd = 0x10 // TX Request
	TxReqExpl      xbeeCmd = 0x11 // Explicit TX Request
	RemoteCmdReq   xbeeCmd = 0x17 // Remote Command Request
	// commands recieved from the xbee

	ATCmdResponse xbeeCmd = 0x88 // AT Command Response
	ModemStatus   xbeeCmd = 0x8A // Modem Status
	TxStatus      xbeeCmd = 0x8B // Transmit Status
	RouteInfoPkt  xbeeCmd = 0x8D // Route information packet
	AddrUpdate    xbeeCmd = 0x8E // Aggregate Addressing Update
	RxPkt         xbeeCmd = 0x90 // RX Indicator (AO=0)
	RxPktExpl     xbeeCmd = 0x91 // Explicit RX Indicator (AO=1)
	IOSample      xbeeCmd = 0x92 // Data Sample RX Indicator
	NodeId        xbeeCmd = 0x95 // Note Identification Indicator
	RemoteCmdResp xbeeCmd = 0x97 // Remote Command Response
)
