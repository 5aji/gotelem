package gotelem

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kschamplin/gotelem/internal/logparsers"
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

// exampleData is a telemetry log data snippet that
// we use to seed the database.
const exampleData = `1698013005.164 1455ED8FDBDFF4FC3BD
1698013005.168 1460000000000000000
1698013005.170 1470000000000000000
1698013005.172 1610000000000000000
1698013005.175 1210000000000000000
1698013005.177 157FFFFC74200000000
1698013005.181 1030000000000000000
1698013005.184 1430000000000000000
1698013005.187 04020D281405EA8FB41
1698013005.210 0413BDF81406AF70042
1698013005.212 042569F81408EF0FF41
1698013005.215 04358A8814041060242
1698013005.219 04481958140D2A40342
1698013005.221 0452DB2814042990442
1698013005.224 047AF948140C031FD41
1698013005.226 04B27A081401ACD0B42
1698013005.229 04DCEAA81403C8C0A42
1698013005.283 04E0378814024580142
1698013005.286 04F97908140BFBC0142
1698013005.289 050098A81402F0F0A42
1698013005.293 051E6AE81402AF20842
1698013005.297 0521AC081403A970742
1698013005.300 0535BB181403CEB0542
1698013005.304 054ECC0814088FE0142
1698013005.307 0554ED181401F44F341
1698013005.309 05726E48140D42BEB41
1698013005.312 059EFC98140EC400142
`

// MakeMockDatabase creates a new dummy database.
func MakeMockDatabase(name string) *TelemDb {
	fstring := fmt.Sprintf("file:%s?mode=memory&cache=shared", name)
	tdb, err := OpenRawDb(fstring)
	if err != nil {
		panic(err)
	}
	// seed the database now.
	scanner := bufio.NewScanner(strings.NewReader(exampleData))

	for scanner.Scan() {
		str := scanner.Text()

		bev, err := logparsers.ParsersMap["telem"](str)
		if err != nil {
			panic(err)
		}
		_, err = tdb.AddEvents(bev)
		if err != nil {
			panic(err)
		}
	}

	return tdb
}

func TestTelemDb(t *testing.T) {


	t.Run("test opening database", func(t *testing.T) {
		// create our mock
		tdb := MakeMockDatabase(t.Name())
		tdb.db.Ping()
	})

	t.Run("test inserting bus event", func(t *testing.T) {
		tdb := MakeMockDatabase(t.Name())
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
			{
				name: "add multiple packet",
				args: args{
					events: []skylab.BusEvent{GetRandomBusEvent(), GetRandomBusEvent()},
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

	t.Run("test getting packets", func(t *testing.T) {
		tdb := MakeMockDatabase(t.Name())

		ctx := context.Background()
		f := BusEventFilter{}
		limitMod := LimitOffsetModifier{Limit: 1}
		pkt, err := tdb.GetPackets(ctx, f, limitMod)
		if err != nil {
			t.Fatalf("error getting packets: %v", err)
		}
		if len(pkt) != 1 {
			t.Fatalf("expected exactly one response, got %d", len(pkt))
		}
		// todo - validate what this should be.
	})

	t.Run("test read-write packet", func(t *testing.T) {
		
	})
}
