package xbee

import (
	"reflect"
	"testing"
)

func TestParseRxFrame(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *RxFrame
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "64-bit unicast",
			args: args{
				// This was taken from the xbee 900hp data sheet, pg 154.
				// the bolded "frame" there has an error and doesn't include the last two bytes.
				data: []byte{0x90, 0x00, 0x13, 0xA2, 0x00, 0x41, 0xAE, 0xB5, 0x4E, 0xFF, 0xFE, 0xC1, 0x54, 0x78, 0x44, 0x61, 0x74, 0x61},
			},
			want: &RxFrame{
				Source:  0x0013A20041AEB54E,
				ACK:     true,
				BCast:   false,
				Payload: []byte{0x54, 0x78, 0x44, 0x61, 0x74, 0x61},
			},
			// Todo: use XCTU to generate more example packets.
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRxFrame(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRxFrame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseRxFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}
