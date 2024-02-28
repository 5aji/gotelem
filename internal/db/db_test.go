package db

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kschamplin/gotelem/skylab"
)

// helper func to get a random bus event with random data.
func GetRandomBusEvent() skylab.BusEvent {
	data := skylab.WsrVelocity{
		MotorVelocity:   1.0,
		VehicleVelocity: 4.0,
	}
	ev := skylab.BusEvent{
		Timestamp: time.Now(),
		Data:      &data,
	}

	return ev
}

func TestTelemDb(t *testing.T) {

	var tdb *TelemDb

	t.Run("test opening database", func(t *testing.T) {
		var err error
		// we use the underlying raw database to avoid the options.
		tdb, err = openRawDb("file::memory:?cache=shared")
		if err != nil {
			t.Errorf("could not open db: %v", err)
		}
		tdb.db.Ping()
		res, _ := tdb.db.Query("SELECT name FROM sqlite_master WHERE type='table'")

		var table string

		for res.Next() {
			res.Scan(&table)
			fmt.Println(table)
		}
	})

	t.Run("test inserting bus event", func(t *testing.T) {
		type args struct {
			events []skylab.BusEvent
		}
		tests := []struct {
			name    string
			args    args
			wantErr bool
		}{
			{
				name: "add no packet",
				args: args{
					events: []skylab.BusEvent{},
				},
				wantErr: false,
			},
			{
				name: "add single packet",
				args: args{
					events: []skylab.BusEvent{GetRandomBusEvent()},
				},
				wantErr: false,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if _, err := tdb.AddEvents(tt.args.events...); (err != nil) != tt.wantErr {
					t.Errorf("TelemDb.AddEvents() error = %v, wantErr %v", err, tt.wantErr)
				}
			})
		}

	})
	type fields struct {
		db *sqlx.DB
	}
	type args struct {
		limit int
		where []QueryFrag
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantEvents []skylab.BusEvent
		wantErr    bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tdb := &TelemDb{
				db: tt.fields.db,
			}
			gotEvents, err := tdb.GetEvents(tt.args.limit, tt.args.where...)
			if (err != nil) != tt.wantErr {
				t.Errorf("TelemDb.GetEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotEvents, tt.wantEvents) {
				t.Errorf("TelemDb.GetEvents() = %v, want %v", gotEvents, tt.wantEvents)
			}
		})
	}
}
