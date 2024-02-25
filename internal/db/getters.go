package db

import (
	"context"
	"fmt"
	"time"
)

// Modifier augments SQL strings.
type Modifier interface {
	ModifyStatement(string) string
}


type LimitOffsetModifier struct {
	Limit int
	Offset int
}

// BusEventFilter is a filter for bus events.
type BusEventFilter struct {
	Names []string
	TimerangeStart time.Time
	TimerangeEnd time.Time
}

// now we can optionally add a limit.

// Datum is a single measurement - it is more granular than a packet.
// the classic example is bms_measurement.current
type Datum struct {
	Timestamp time.Time `db:"timestamp"`
	Value     any       `db:"val"`
}

// GetValues queries the database for values in a given time range.
// A value is a specific data point. For example, bms_measurement.current
// would be a value.
func (tdb *TelemDb) GetValues(ctx context.Context, packetName, field string, start time.Time,
	end time.Time) ([]Datum, error) {
	// this fragment uses json_extract from sqlite to get a single
	// nested value.
	SqlFrag := `
	SELECT 
	ts as timestamp,
	json_extract(data, '$.' || ?) as val
	FROM bus_events WHERE name IS ? 
	`
	rows, err := tdb.db.QueryxContext(ctx, SqlFrag, field, packetName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := make([]Datum, 0, 10)
	for rows.Next() {
		var d Datum = Datum{}
		err = rows.StructScan(&d)
		if err != nil {
			fmt.Print(err)
			return data, err
		}
		data = append(data, d)
	}
	fmt.Print(rows.Err())

	return data, nil
}
