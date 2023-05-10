package xbee

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"
)

func Test_xbeeFrameSplit(t *testing.T) {
	advTest, _ := hex.DecodeString("007E7E0012900013A20041B320")
	type args struct {
		data  []byte
		atEOF bool
	}
	tests := []struct {
		name        string
		args        args
		wantAdvance int
		wantToken   []byte
		wantErr     bool
	}{
		{
			name: "empty data",
			args: args{
				data:  []byte{},
				atEOF: false,
			},
			wantAdvance: 0,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name: "EOF and empty data",
			args: args{
				data:  []byte{},
				atEOF: true,
			},
			wantAdvance: 0,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name: "no start delimiter",
			args: args{
				data:  []byte{0x11, 0x22, 0x23, 0x44, 0x44, 0x77, 0x33},
				atEOF: false,
			},
			wantAdvance: 7,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name: "incomplete packet",
			args: args{
				data:  []byte{0x7E, 0x00, 0x02, 0x23, 0x11},
				atEOF: false,
			},
			wantAdvance: 0,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name: "valid packet",
			args: args{
				data:  []byte{0x7E, 0x00, 0x02, 0x23, 0x11, 0xCB},
				atEOF: false,
			},
			wantAdvance: 6,
			wantToken:   []byte{0x7E, 0x00, 0x02, 0x23, 0x11, 0xCB},
			wantErr:     false,
		},
		{
			name: "valid packet w/ padding",
			args: args{
				data:  []byte{0x00, 0x7E, 0x00, 0x02, 0x23, 0x11, 0xCB, 0x00},
				atEOF: false,
			},
			wantAdvance: 7,
			wantToken:   []byte{0x7E, 0x00, 0x02, 0x23, 0x11, 0xCB},
			wantErr:     false,
		},
		{
			name: "trailing start delimiter",
			args: args{
				data:  []byte{0x53, 0x00, 0x02, 0x23, 0x11, 0x7E},
				atEOF: false,
			},
			wantAdvance: 5,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name: "incomplete length value",
			args: args{
				data:  []byte{0x53, 0x00, 0x02, 0x23, 0x11, 0x7E, 0x00},
				atEOF: false,
			},
			wantAdvance: 5,
			wantToken:   nil,
			wantErr:     false,
		},
		{
			name: "start delimiter inside partial packet",
			args: args{
				data: advTest,
				atEOF: false,
			},
			wantAdvance: 2,
			wantToken: nil,
			wantErr: false,
		},

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAdvance, gotToken, err := xbeeFrameSplit(tt.args.data, tt.args.atEOF)
			if (err != nil) != tt.wantErr {
				t.Errorf("xbeeFrameSplit() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotAdvance != tt.wantAdvance {
				t.Errorf("xbeeFrameSplit() gotAdvance = %v, want %v", gotAdvance, tt.wantAdvance)
			}
			if !reflect.DeepEqual(gotToken, tt.wantToken) {
				t.Errorf("xbeeFrameSplit() gotToken = %v, want %v", gotToken, tt.wantToken)
			}
		})
	}
}



func Test_parseFrame(t *testing.T) {
	type args struct {
		frame []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "missing start delimiter",
			args: args{
				frame: []byte{0x00, 0x02, 0x03, 0x00, 0x3},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "checksum mismatch",
			args: args{
				frame: []byte{0x7E, 0x00, 0x02, 0x23, 0x11, 0xCA},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "valid packet",
			args: args{
				frame: []byte{0x7E, 0x00, 0x02, 0x23, 0x11, 0xCB},
			},
			want:    []byte{0x23, 0x11},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFrame(tt.args.frame)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseFrame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_writeXBeeFrame(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		wantN   int
		wantW   []byte
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "a test",
			args: args{
				data: []byte{0x23, 0x11},
			},
			wantN:   6,
			wantW:   []byte{0x7E, 0x00, 0x02, 0x23, 0x11, 0xCB},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			gotN, err := writeXBeeFrame(w, tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("writeXBeeFrame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotN != tt.wantN {
				t.Errorf("writeXBeeFrame() = %v, want %v", gotN, tt.wantN)
			}
			if gotW := w.Bytes(); !reflect.DeepEqual(gotW, tt.wantW) {
				t.Errorf("writeXBeeFrame() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
