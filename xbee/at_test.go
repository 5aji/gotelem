package xbee

import (
	"reflect"
	"testing"
)

func TestParseATCmdResponse(t *testing.T) {
	type args struct {
		p []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *ATCmdResponse
		wantErr bool
	}{
		{
			name: "AT Command Change OK",
			args: args{
				p: []byte{0x88, 0x53, 0x49, 0x44, 0x00},
			},
			want: &ATCmdResponse{
				Cmd:    "ID",
				Data:   []byte{},
				Status: ATCmdStatusOK,
			},
			wantErr: false,
		},
		{
			name: "AT Command Query OK",
			args: args{
				p: []byte{0x88, 0x53, 0x49, 0x44, 0x00, 0x43, 0xEF},
			},
			want: &ATCmdResponse{
				Cmd:    "ID",
				Data:   []byte{0x43, 0xEF},
				Status: ATCmdStatusOK,
			},
		},
		{
			name: "AT Command Parameter Error",
			args: args{
				p: []byte{0x88, 0x53, 0x49, 0x44, 0x03},
			},
			want: &ATCmdResponse{
				Cmd:    "ID",
				Data:   []byte{},
				Status: ATCmdStatusInvalidParam,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseATCmdResponse(tt.args.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseATCmdResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseATCmdResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_encodeRemoteATCommand(t *testing.T) {
	type args struct {
		at          ATCmd
		idx         uint8
		queued      bool
		destination uint64
	}
	tests := []struct {
		name string
		args args
		want RawATCmd
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeRemoteATCommand(tt.args.at, tt.args.idx, tt.args.queued, tt.args.destination); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeRemoteATCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_encodeATCommand(t *testing.T) {
	type args struct {
		cmd    [2]rune
		p      []byte
		idx    uint8
		queued bool
	}
	tests := []struct {
		name string
		args args
		want RawATCmd
	}{
		// These test cases are from digi's documentation on the 900HP/XSC modules.
		{
			name: "Setting AT Command",
			args: args{
				cmd:    [2]rune{'N', 'I'},
				idx:    0xA1,
				p:      []byte{0x45, 0x6E, 0x64, 0x20, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65},
				queued: false,
			},
			want: []byte{0x08, 0xA1, 0x4E, 0x49, 0x45, 0x6E, 0x64, 0x20, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65},
		},
		{
			name: "Query AT Command",
			args: args{
				cmd:    [2]rune{'T', 'P'},
				idx:    0x17,
				p:      nil,
				queued: false,
			},
			want: []byte{0x08, 0x17, 0x54, 0x50},
		},
		{
			name: "Queue Local AT Command",
			args: args{
				cmd:    [2]rune{'B', 'D'},
				idx:    0x53,
				p:      []byte{0x07},
				queued: true,
			},
			want: []byte{0x09, 0x53, 0x42, 0x44, 0x07},
		},
		{
			name: "Queue Query AT Command",
			args: args{
				cmd:    [2]rune{'T', 'P'},
				idx:    0x17,
				p:      nil,
				queued: true,
			},
			want: []byte{0x09, 0x17, 0x54, 0x50},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := encodeATCommand(tt.args.cmd, tt.args.p, tt.args.idx, tt.args.queued); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("encodeATCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
