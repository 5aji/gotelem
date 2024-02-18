package db

import (
	"context"
	"fmt"
	"time"
)

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
	const SqlFrag = `
	SELECT 
	datetime(ts /1000.0, 'unixepoch', 'subsec') as timestamp,
	json_extract(data, '$.' || ?) as val
	FROM bus_events WHERE name IS ? AND timestamp IS NOT NULL
	`
	fmt.Print(start, end, packetName, field)

	rows, err := tdb.db.QueryxContext(ctx, SqlFrag, field, packetName, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := make([]Datum, 0, 10)
	for rows.Next() {
		var d Datum
		err = rows.StructScan(&d)
		if err != nil {
			fmt.Print(err)
			return data, err
		}
		data = append(data, d)
	}
	fmt.Print(data)

	return data, nil
}
