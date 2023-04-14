package can

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"golang.org/x/sys/unix"
)

// this file implements a simple wrapper around linux socketCAN
type CanSocket struct {
	iface *net.Interface
	addr  *unix.SockaddrCAN
	fd    int
}

type Frame struct {
	ID   uint32
	Data []uint8
	Kind Kind
}

//internal frame structure for socketcan with padding

type stdFrame struct {
	ID          uint32
	Len         uint8
	_pad, _res1 uint8 // padding
	Dlc         uint8
	Data        [8]uint8
}

func Marshal(f Frame) (*bytes.Buffer, error) {

	if len(f.Data) > 8 && f.Kind == SFF {
		return nil, errors.New("data too large for std frame")
	}
	if len(f.Data) > 64 && f.Kind == EFF {
		return nil, errors.New("data too large for extended frame")
	}

	var idflags uint32 = f.ID
	if f.Kind == EFF {
		idflags = idflags | unix.CAN_EFF_FLAG
	} else if f.Kind == RTR {
		idflags = idflags | unix.CAN_RTR_FLAG
	} else if f.Kind == ERR {
		idflags = idflags | unix.CAN_ERR_FLAG
	}

	var d [8]uint8

	for i := 0; i < len(f.Data); i++ {
		d[i] = f.Data[i]
	}

	var unixFrame stdFrame = stdFrame{
		ID: idflags, Len: uint8(len(f.Data)),
		Data: d,
	}

	// finally, write our bytes buffer.
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, unixFrame)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

//go:generate stringer -output=frame_kind.go -type Kind
type Kind uint8

const (
	SFF Kind = iota // Standard Frame Format
	EFF             // Extended Frame
	RTR             // remote transmission requests
	ERR             // Error frame.
)

// helper function to make a filter.
// id and mask are straightforward, if inverted is true, the filter
// will reject anything that matches.
func MakeFilter(id, mask uint32, inverted bool) *unix.CanFilter {
	f := &unix.CanFilter{Id: id, Mask: mask}

	if inverted {
		f.Id = f.Id | unix.CAN_INV_FILTER
	}
	return f
}

const standardFrameSize = unix.CAN_MTU

// the hack.
const extendedFrameSize = unix.CAN_MTU + 56

// Constructs a new CanSocket and creates a file descriptor for it.
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

// close the socket file descriptor, freeing it from the system.
func (sck *CanSocket) Close() error {
	return unix.Close(sck.fd)
}

// get the name of the socket, or nil if it hasn't been bound yet.
func (sck *CanSocket) Name() string {
	return sck.iface.Name
}

// should we log errors?
func (sck *CanSocket) SetErrFilter(shouldFilter bool) error {

	var err error
	var errmask = 0
	if shouldFilter {
		errmask = unix.CAN_ERR_MASK
	}

	err = unix.SetsockoptInt(sck.fd, unix.SOL_CAN_RAW, unix.CAN_RAW_ERR_FILTER, errmask)
	if err != nil {
		return err
	}
	return nil
}

func (sck *CanSocket) SetFilters(filters []unix.CanFilter) error {
	return unix.SetsockoptCanRawFilter(sck.fd, unix.SOL_CAN_RAW, unix.CAN_RAW_FILTER, filters)

}

func (sck *CanSocket) Send(msg Frame) error {
	// convert our abstract frame into a real unix frame and then push it.
	// check return value to raise errors.
	buf, err := Marshal(msg)

	if err != nil {
		return fmt.Errorf("error sending frame: %w", err)
	}

	if buf.Len() != unix.CAN_MTU {
		return fmt.Errorf("socket send: buffer size mismatch %d", buf.Len())
	}
	// send the buffer using unix syscalls!

	err = unix.Send(sck.fd, buf.Bytes(), 0)
	if err != nil {
		return fmt.Errorf("error sending frame: %w", err)
	}

	return nil
}

func (sck *CanSocket) Recv() (*Frame, error) {
	return nil, errors.New("not implemented")
}
