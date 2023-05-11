/*
Package xbee provides communication and configuration of Digi XBee products

(and other Digi products that are similar such as the XLR Pro). It provides
a net.Conn-like interface as well as AT commands for configuration. The most
common usage of the package is with a Session, which provides
*/
package xbee

import (
	"encoding/binary"
	"os"
	"reflect"
	"testing"

	"golang.org/x/exp/slog"
)

func TestXBeeHardware(t *testing.T) {
	// this test runs only if the environemnt variable
	// XBEE_DEVICE is set.

	devStr, ok := os.LookupEnv("XBEE_DEVICE")

	if !ok {
		t.Skip("No XBee device provided")
	}

	var sess *Session = nil
	// test connection.
	t.Run("Connect to device", func(t *testing.T) {
		dev, err := ParseDeviceString(devStr)
		if err != nil {
			t.Errorf("ParseDeviceString() error = %v", err)
		}
		sess, err = NewSession(dev, slog.With("type", dev.Type()))
		if err != nil {
			t.Errorf("NewSession() error = %v", err)
		}
		// err = sess.Close()
		// if err != nil {
		// 	t.Errorf("Session.Close() error = %v", err)
		// }

	})

	// now we should test sending a packet. and getting a response.

	t.Run("Get Network ID", func(t *testing.T) {
		b, err := sess.ATCommand([2]rune{'I', 'D'}, nil, false)
		if err != nil {
			t.Errorf("ATCommand() error = %v", err)
		}
		if len(b) != 2 {
			t.Errorf("reponse length mismatch: expected 2 got %d", len(b))
		}
	})

	t.Run("Check NP", func(t *testing.T) {
		b, err := sess.ATCommand([2]rune{'N', 'P'}, nil, false)

		if err != nil {
			t.Errorf("ATCommand() error = %v", err)
		}

		val := binary.BigEndian.Uint16(b)
		if val != 0x100 && val != 0x640 {
			t.Errorf("NP response wrong, expected 0x100 or 0x640 got 0x%X", val)
		}
	})

}

func TestParseDeviceString(t *testing.T) {
	type args struct {
		dev string
	}
	tests := []struct {
		name    string
		args    args
		want    *Transport
		wantErr bool
	}{
		{
			name: "invalid stuff",
			args: args{
				dev: "blah",
			},
			want: nil,
			wantErr: true,
		},
		// TODO: moar tests!
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDeviceString(tt.args.dev)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDeviceString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseDeviceString() = %v, want %v", got, tt.want)
			}
		})
	}
}
