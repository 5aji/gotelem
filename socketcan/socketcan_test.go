//go:build linux

package socketcan

import (
	"bytes"
	"net"
	"testing"

	"github.com/kschamplin/gotelem"
)

func TestCanSocket(t *testing.T) {

	if _, err := net.InterfaceByName("vcan0"); err != nil {
		t.Skipf("missing vcan0, skipping socket tests: %v", err)
	}

	t.Run("test construction and destruction", func(t *testing.T) {
		sock, err := NewCanSocket("vcan0")
		if err != nil {
			t.Errorf("could not make socket: %v", err)
		}
		if sock.fd == 0 {
			t.Errorf("socket was not made: expected non-zero, got %d", sock.fd)
		}
		if err := sock.Close(); err != nil {
			t.Errorf("could not close socket")
		}
	})

	t.Run("test name", func(t *testing.T) {
		sock, _ := NewCanSocket("vcan0")
		defer sock.Close()

		if sock.Name() != "vcan0" {
			t.Errorf("incorrect interface name: got %s, expected %s", sock.Name(), "vcan0")
		}
	})

	t.Run("test sending can 2.0 packet", func(t *testing.T) {
		sock, _ := NewCanSocket("vcan0")
		defer sock.Close()

		// make a packet.
		testFrame := &gotelem.Frame{
			Id:   0x123,
			Kind: gotelem.SFF,
			Data: []byte{0, 1, 2, 3, 4, 5, 6, 7},
		}
		err := sock.Send(testFrame)

		if err != nil {
			t.Error(err)
		}
	})

	t.Run("test receiving a can 2.0 packet", func(t *testing.T) {
		sock, _ := NewCanSocket("vcan0")
		rsock, _ := NewCanSocket("vcan0")
		defer sock.Close()
		defer rsock.Close()

		testFrame := &gotelem.Frame{
			Id:   0x234,
			Kind: gotelem.SFF,
			Data: []byte{0, 1, 2, 3, 4, 5, 6, 7},
		}
		_ = sock.Send(testFrame)

		rpkt, err := rsock.Recv()
		if err != nil {
			t.Error(err)
		}
		if len(rpkt.Data) != 8 {
			t.Errorf("length mismatch: got %d expected 8", len(rpkt.Data))
		}
		if !bytes.Equal(testFrame.Data, rpkt.Data) {
			t.Error("data corrupted")
		}

	})

}
