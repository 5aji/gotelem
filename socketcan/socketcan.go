//go:build linux

/*
Package socketcan provides a wrapper around the Linux socketCAN interface.
*/
package socketcan

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/kschamplin/gotelem"
	"golang.org/x/sys/unix"
)

// A CanSocket is a CAN device that uses the socketCAN linux drivers to write to real
// CAN hardware.
type CanSocket struct {
	iface *net.Interface
	addr  *unix.SockaddrCAN
	fd    int
}

// CanFilter is a filter for an interface.
type CanFilter interface {
	Inverted() bool
	Mask() uint32
	Id() uint32
}

// standardFrameSize is the full size in bytes of the default CAN frame.
const standardFrameSize = unix.CAN_MTU

// we use the base CAN_MTU since the FD MTU is not in sys/unix. but we know it's +64-8 bytes
const fdFrameSize = unix.CAN_MTU + 56

// NewCanSocket constructs a new CanSocket and binds it to the interface given by ifname
func NewCanSocket(ifname string) (*CanSocket, error) {

	var sck CanSocket
	fd, err := unix.Socket(unix.AF_CAN, unix.SOCK_RAW, unix.CAN_RAW)
	if err != nil {
		return nil, err
	}

	sck.fd = fd

	iface, err := net.InterfaceByName(ifname)

	if err != nil {
		return nil, err
	}

	sck.iface = iface

	sck.addr = &unix.SockaddrCAN{Ifindex: sck.iface.Index}

	err = unix.Bind(sck.fd, sck.addr)
	if err != nil {
		return nil, err
	}

	return &sck, nil
}

// Close closes the socket.
func (sck *CanSocket) Close() error {
	return unix.Close(sck.fd)
}

// Name returns the name of the socket.
func (sck *CanSocket) Name() string {
	return sck.iface.Name
}

// SetErrFilter sets if error packets should be sent upstream
func (sck *CanSocket) SetErrFilter(shouldFilter bool) error {

	var err error
	var errmask = 0
	if shouldFilter {
		errmask = unix.CAN_ERR_MASK
	}

	err = unix.SetsockoptInt(sck.fd, unix.SOL_CAN_RAW, unix.CAN_RAW_ERR_FILTER, errmask)

	return err
}

// SetFDMode enables or disables the transmission of CAN FD packets.
func (sck *CanSocket) SetFDMode(enable bool) error {
	var val int
	if enable {
		val = 1
	} else {
		val = 0
	}

	err := unix.SetsockoptInt(sck.fd, unix.SOL_CAN_RAW, unix.CAN_RAW_FD_FRAMES, val)

	return err

}

// SetFilters will set the socketCAN filters based on a standard CAN filter list.
func (sck *CanSocket) SetFilters(filters []CanFilter) error {

	// helper function to make a filter.
	// id and mask are straightforward, if inverted is true, the filter
	// will reject anything that matches.
	makeFilter := func(filter CanFilter) unix.CanFilter {
		f := unix.CanFilter{Id: filter.Id(), Mask: filter.Mask()}

		if filter.Inverted() {
			f.Id = f.Id | unix.CAN_INV_FILTER
		}
		return f
	}

	convertedFilters := make([]unix.CanFilter, len(filters))
	for i, filt := range filters {
		convertedFilters[i] = makeFilter(filt)
	}
	return unix.SetsockoptCanRawFilter(sck.fd, unix.SOL_CAN_RAW, unix.CAN_RAW_FILTER, convertedFilters)

}

// Send sends a CAN frame
func (sck *CanSocket) Send(msg *gotelem.Frame) error {

	buf := make([]byte, fdFrameSize)

	idToWrite := msg.Id

	switch msg.Kind {
	case gotelem.CanSFFFrame:
		idToWrite &= unix.CAN_SFF_MASK
	case gotelem.CanEFFFrame:
		idToWrite &= unix.CAN_EFF_MASK
		idToWrite |= unix.CAN_EFF_FLAG
	case gotelem.CanRTRFrame:
		idToWrite |= unix.CAN_RTR_FLAG
	case gotelem.CanErrFrame:
		return errors.New("you can't send error frames")
	default:
		return errors.New("unknown frame type")
	}

	binary.LittleEndian.PutUint32(buf[:4], idToWrite)

	// write the length, it's one byte, so do it directly.
	payloadLength := len(msg.Data)
	buf[4] = byte(payloadLength)

	if payloadLength > 64 {
		return fmt.Errorf("payload too large: %d", payloadLength)
	}

	// copy in the data now.
	copy(buf[8:], msg.Data)

	// send the buffer using unix syscalls!
	var err error
	if payloadLength > 8 {
		err = unix.Send(sck.fd, buf, 0)
	} else {
		err = unix.Send(sck.fd, buf[:standardFrameSize], 0)
	}
	if err != nil {
		return fmt.Errorf("error sending frame: %w", err)
	}

	return nil
}

func (sck *CanSocket) Recv() (*gotelem.Frame, error) {

	// todo: support extended frames.
	buf := make([]byte, fdFrameSize)
	_, err := unix.Read(sck.fd, buf)
	if err != nil {
		return nil, err
	}

	id := binary.LittleEndian.Uint32(buf[0:4])

	var k gotelem.Kind
	if id&unix.CAN_EFF_FLAG != 0 {
		// extended id frame
		k = gotelem.CanEFFFrame
	} else {
		// it's a normal can frame
		k = gotelem.CanSFFFrame
	}

	dataLength := uint8(buf[4])

	result := &gotelem.Frame{
		Id:   id & unix.CAN_EFF_MASK,
		Kind: k,
		Data: buf[8 : dataLength+8],
	}
	return result, nil

}
