package db

import (
	"embed"
	"reflect"
	"testing"
)

//go:embed migrations/1_*.sql
//go:embed migrations/2_*.sql
var testFs embed.FS

func Test_getMigrations(t *testing.T) {
	tests := []struct {
		name string
		want map[int]map[string]Migration
	}{
		{
			name: "main test",
			want: map[int]map[string]Migration{
				1: {
					"up": Migration{
						Name:     "initial",
						Version:  1,
						FileName: "1_initial_up.sql",
					},
					"down": Migration{
						Name:     "initial",
						Version:  1,
						FileName: "1_initial_down.sql",
					},
				},

				2: {
					"up": Migration{
						Name:     "addl_tables",
						Version:  2,
						FileName: "2_addl_tables_up.sql",
					},
					"down": Migration{
						Name:     "addl_tables",
						Version:  2,
						FileName: "2_addl_tables_down.sql",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMigrations(testFs); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMigrations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRunMigrations(t *testing.T) {
	type args struct {
		tdb *TelemDb
	}
	tests := []struct {
		name         string
		args         args
		wantFinalVer int
		wantErr      bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFinalVer, err := RunMigrations(tt.args.tdb)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunMigrations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotFinalVer != tt.wantFinalVer {
				t.Errorf("RunMigrations() = %v, want %v", gotFinalVer, tt.wantFinalVer)
			}
		})
	}
}
