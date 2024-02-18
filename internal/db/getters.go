package db

import (
	"context"
	"strings"
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
func (tdb *TelemDb) GetValues(ctx context.Context, packetName, field string , start time.Time,
	end time.Time) ([]Datum, error) {
	// this fragment uses json_extract from sqlite to get a single
	// nested value.
	const SqlFrag = `
	SELECT 
	datetime(ts /1000.0, 'unixepoch', 'subsec') as timestamp,
	json_extract(data, ?) as val,
	FROM bus_events WHERE name IS ? AND timestamp BETWEEN ? AND ?
	`

	fieldJson := "$." + field

	rows, err := tdb.db.QueryxContext(ctx, fieldJson, packetName, start, end)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	data := make([]Datum, 0, 10)
	for rows.Next() {
		var d Datum
		err = rows.StructScan(&d)
		if err != nil {
			return data, err
		}
		data = append(data, d)
	}

	return data, nil
}

