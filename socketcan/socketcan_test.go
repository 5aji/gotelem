package socketcan

import (
	"testing"

	"golang.org/x/sys/unix"
)

func TestMakeFilter(t *testing.T) {

	t.Run("non-invert", func(t *testing.T) {

		filter := MakeFilter(0x123, 0x11, false)

		if filter.Id != 0x123 {
			t.Errorf("expected %d, got %d", 0x123, filter.Id)
		}
		if filter.Mask != 0x11 {
			t.Errorf("expected %d, got %d", 0x11, filter.Mask)
		}
	})
	t.Run("invert", func(t *testing.T) {

		filter := MakeFilter(0x123, 0x11, true)

		if filter.Id != 0x123|unix.CAN_INV_FILTER {
			t.Errorf("expected %d, got %d", 0x123|unix.CAN_INV_FILTER, filter.Id)
		}
		if filter.Mask != 0x11 {
			t.Errorf("expected %d, got %d", 0x11, filter.Mask)
		}
	})
}

func TestCanSocket(t *testing.T) {

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

		if sock.Name() != "vcan0" {
			t.Errorf("incorrect interface name: got %s, expected %s", sock.Name(), "vcan0")
		}
	})

}
