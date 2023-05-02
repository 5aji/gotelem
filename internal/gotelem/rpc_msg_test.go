package gotelem

import (
	"reflect"
	"testing"

	"github.com/tinylib/msgp/msgp"
)

func Test_parseRPC(t *testing.T) {
	type args struct {
		raw msgp.Raw
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRPC(tt.args.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseRPC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseRPC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getMsgType(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want RPCType
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMsgType(tt.args.b); got != tt.want {
				t.Errorf("getMsgType() = %v, want %v", got, tt.want)
			}
		})
	}
}
