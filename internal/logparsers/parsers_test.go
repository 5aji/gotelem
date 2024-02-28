package logparsers

import (
	"reflect"
	"testing"
	"time"

	"github.com/kschamplin/gotelem/internal/can"
)

func Test_parseCanDumpLine(t *testing.T) {
	type args struct {
		dumpLine string
	}
	tests := []struct {
		name      string
		args      args
		wantFrame can.Frame
		wantTs    time.Time
		wantErr   bool
	}{
		{
			name: "test garbage",
			args: args{dumpLine: "hosireoie"},
			wantFrame: can.Frame{},
			wantTs: time.Unix(0,0),
			wantErr: true,
		},
		{
			name: "test normal data",
			args: args{dumpLine: "(1684538768.521889) can0 200#8D643546"},
			wantFrame: can.Frame{
				Id: can.CanID{Id: 0x200, Extended: false},
				Data: []byte{0x8d, 0x64, 0x35, 0x46},
				Kind: can.CanDataFrame,
			},
			wantTs: time.Unix(1684538768, 521889),
			wantErr: false,
		},
		{
			name: "bad data length",
			// odd number of hex data nibbles
			args: args{dumpLine: "(1684538768.521889) can0 200#8D64354"},
			wantFrame: can.Frame{},
			wantTs: time.Unix(0,0),
			wantErr: true,
		},
		{
			name: "invalid hex",
			// J is not valid hex.
			args: args{dumpLine: "(1684538768.521889) can0 200#8D64354J"},
			wantFrame: can.Frame{},
			wantTs: time.Unix(0,0),
			wantErr: true,
		},
		{
			name: "bad time",
			// we destroy the time structure.
			args: args{dumpLine: "(badtime.521889) can0 200#8D643546"},
			wantFrame: can.Frame{},
			wantTs: time.Unix(0,0),
			wantErr: true,
		},
		// TODO: add extended id test case

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrame, gotTs, err := parseCanDumpLine(tt.args.dumpLine)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCanDumpLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFrame, tt.wantFrame) {
				t.Errorf("parseCanDumpLine() gotFrame = %v, want %v", gotFrame, tt.wantFrame)
			}
			if !reflect.DeepEqual(gotTs, tt.wantTs) {
				t.Errorf("parseCanDumpLine() gotTs = %v, want %v", gotTs, tt.wantTs)
			}
		})
	}
}
