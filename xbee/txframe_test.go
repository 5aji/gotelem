package xbee

import (
	"reflect"
	"testing"
)

func TestTxFrame_Bytes(t *testing.T) {
	type fields struct {
		Id          byte
		Destination uint64
		BCastRadius uint8
		Options     uint8
		Payload     []byte
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		// TODO: Add test cases.
		{
			name: "64 bit unicast",
			fields: fields{
				Id:          0x52,
				Destination: 0x0013A20012345678,
				BCastRadius: 0,
				Options:     0,
				Payload:     []byte{0x54, 0x78, 0x44, 0x61, 0x74, 0x61},
			},
			want: []byte{0x10, 0x52, 0x00, 0x13, 0xA2, 0x00, 0x12, 0x34, 0x56, 0x78, 0xFF, 0xFE, 0x00, 0x00, 0x54, 0x78, 0x44, 0x61, 0x74, 0x61},
		},
		{
			name: "64 bit broadcast",
			fields: fields{
				Id:          0x00,
				Destination: 0xFFFF,
				BCastRadius: 1,
				Options:     0,
				Payload:     []byte{0x42, 0x72, 0x6F, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74},
			},
			want: []byte{0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFE, 0x01, 0x00, 0x42, 0x72, 0x6F, 0x61, 0x64, 0x63, 0x61, 0x73, 0x74},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			txFrame := &TxFrame{
				Id:          tt.fields.Id,
				Destination: tt.fields.Destination,
				BCastRadius: tt.fields.BCastRadius,
				Options:     tt.fields.Options,
				Payload:     tt.fields.Payload,
			}
			if got := txFrame.Bytes(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TxFrame.Bytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseTxStatusFrame(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *TxStatusFrame
		wantErr bool
	}{
		{
			name: "wrong packet type",
			args: args{
				data: []byte{0x85, 0x47, 0xFF, 0xFE, 0x00, 0x00, 0x02},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "wrong packet length",
			args: args{
				data: []byte{0x8B, 0x47, 0xFF, 0xFE, 0x00, 0x00},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success packet",
			args: args{
				data: []byte{0x8B, 0x47, 0xFF, 0xFE, 0x00, 0x00, 0x00},
			},
			want: &TxStatusFrame{
				Id:     0x47,
				NRetry: 0,
				Status: TxStatusSuccess,
				Routed: false,
			},
			wantErr: false,
		},
		{
			name: "ack fail packet",
			args: args{
				data: []byte{0x8B, 0x47, 0xFF, 0xFE, 0x00, 0x01, 0x00},
			},
			want: &TxStatusFrame{
				Id:     0x47,
				NRetry: 0,
				Status: TxStatusNoACK,
				Routed: false,
			},
			wantErr: false,
		},
		{
			name: "routed retried packet",
			args: args{
				data: []byte{0x8B, 0x47, 0xFF, 0xFE, 0x03, 0x01, 0x02},
			},
			want: &TxStatusFrame{
				Id:     0x47,
				NRetry: 3,
				Status: TxStatusNoACK,
				Routed: true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTxStatusFrame(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTxStatusFrame() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTxStatusFrame() = %v, want %v", got, tt.want)
			}
		})
	}
}
