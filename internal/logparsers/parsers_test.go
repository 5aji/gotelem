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
			name: "test normal data",
			args: args{dumpLine: "(1684538768.521889) can0 200#8D643546"},
			wantFrame: can.Frame{
				Id:   can.CanID{Id: 0x200, Extended: false},
				Data: []byte{0x8d, 0x64, 0x35, 0x46},
				Kind: can.CanDataFrame,
			},
			wantTs:  time.Unix(1684538768, 521889*int64(time.Microsecond)),
			wantErr: false,
		},
		// TODO: add extended id test case

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrame, gotTs, err := parseCanDumpLine(tt.args.dumpLine)
			if (err == nil) == tt.wantErr {
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

func Test_parseCanDumpLine_errors(t *testing.T) {
	// this test tries a bunch of failure cases to ensure that they are caught and not panicking.

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "garbage input",
			input: "hoiseorhijkl",
		},
		{
			name: "bad data length",
			// odd number of hex data nibbles
			input: "(1684538768.521889) can0 200#8D64354",
		},
		{
			name: "invalid hex",
			// J is not valid hex.
			input: "(1684538768.521889) can0 200#8D64354J",
		},
		{
			name: "bad time",
			// we destroy the time structure.
			input: "(badtime.521889) can0 200#8D643546",
		},
		{
			name: "utf8 corruption",
			// we attempt to mess up the data with broken utf8
			input: "(1684538768.521889) can0 200#8D6\xed\xa0\x8043546",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, ts, err := parseCanDumpLine(tt.input)

			if err == nil {
				t.Fatalf("parseCanDumpLine() expected error but instead got f = %v, ts = %v", f, ts)
			}
		})
	}
}

func Test_parseTelemLogLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name      string
		args      args
		wantFrame can.Frame
		wantTs    time.Time
		wantErr   bool
	}{
		{
			name: "basic test",
			args: args{line: "1698180835.318 0619D80564080EBE241"},
			wantFrame: can.Frame{
				Id:   can.CanID{Id: 0x61, Extended: false},
				Data: []byte{0x9D, 0x80, 0x56, 0x40, 0x80, 0xEB, 0xE2, 0x41},
				Kind: can.CanDataFrame,
			},
			wantTs:  time.Unix(1698180835, 318*int64(time.Millisecond)),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFrame, gotTs, err := parseTelemLogLine(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTelemLogLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotFrame, tt.wantFrame) {
				t.Errorf("parseTelemLogLine() gotFrame = %v, want %v", gotFrame, tt.wantFrame)
			}
			if !reflect.DeepEqual(gotTs, tt.wantTs) {
				t.Errorf("parseTelemLogLine() gotTs = %v, want %v", gotTs, tt.wantTs)
			}
		})
	}
}

func Test_parseTelemLogLine_errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "garbage input",
			input: "ajl;ksdoifhge\xEB",
		},
		{
			name:  "bad data length",
			input: "1698180835.318 0619D80564080EBE24",
		},
		{
			name:  "bad timestamp",
			input: "99999999999999999999999999999999999999999999999.318 0619D80564080EBE24",
		},
		{
			name:  "invalid hex characters",
			input: "1698180835.318 0619D805640X0EBE24",
		},
		{
			name:  "utf8 corruption",
			input: "1698180835.318 0619\xed\xa0\x80fsadfD805640X0EBE24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, ts, err := parseTelemLogLine(tt.input)
			if err == nil {
				t.Fatalf("parseTelemLogLine() expected error but instead got f = %v, ts = %v", f, ts)
			}
		})
	}
}
