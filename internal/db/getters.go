package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kschamplin/gotelem/skylab"
)

// Modifier augments SQL strings.
type Modifier interface {
	ModifyStatement(*strings.Builder) error
}

// LimitOffsetModifier is a modifier to support pagniation.
type LimitOffsetModifier struct {
	Limit  int
	Offset int
}

func (l LimitOffsetModifier) ModifyStatement(sb *strings.Builder) error {
	clause := fmt.Sprintf(" LIMIT %d OFFSET %d", l.Limit, l.Offset)
	sb.WriteString(clause)
	return nil
}

// BusEventFilter is a filter for bus events.
type BusEventFilter struct {
	Names          []string
	TimerangeStart time.Time
	TimerangeEnd   time.Time
}

// now we can optionally add a limit.

func (tdb *TelemDb) GetPackets(ctx context.Context, filter BusEventFilter, options ...Modifier) ([]skylab.BusEvent, error) {
	// construct a simple
	var whereFrags = make([]string, 0)

	// if we're filtering by names, add a where clause for it.
	if len(filter.Names) > 0 {
		names := strings.Join(filter.Names, ", ")
		qString := fmt.Sprintf("name IN (%s)", names)
		whereFrags = append(whereFrags, qString)
	}
	// TODO: identify if we need a special case for both time ranges
	// using BETWEEN since apparenlty that can be better?

	// next, check if we have a start/end time, add constraints
	if !filter.TimerangeEnd.IsZero() {
		qString := fmt.Sprintf("ts <= %d", filter.TimerangeEnd.UnixMilli())
		whereFrags = append(whereFrags, qString)
	}
	if !filter.TimerangeStart.IsZero() {
		// we have an end range
		qString := fmt.Sprintf("ts >= %d", filter.TimerangeStart.UnixMilli())
		whereFrags = append(whereFrags, qString)
	}

	sb := strings.Builder{}
	sb.WriteString("SELECT * from \"bus_events\"")
	// construct the full statement.
	if len(whereFrags) > 0 {
		// use the where clauses.
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(whereFrags, " AND "))
	}

	// Augment our data further if there's i.e a limit modifier.
	// TODO: factor this out maybe?
	for _, m := range options {
		m.ModifyStatement(&sb)
	}
	rows, err := tdb.db.QueryxContext(ctx, sb.String())
	if err != nil { 
		return nil, err
	}
	defer rows.Close()

	var events = make([]skylab.BusEvent, 0, 10)

	for rows.Next() {
		var ev skylab.RawJsonEvent
		err := rows.Scan(&ev.Timestamp, &ev.Name, (*[]byte)(&ev.Data))
		if err != nil {
			return nil, err
		}

		BusEv := skylab.BusEvent {
			Timestamp: time.UnixMilli(int64(ev.Timestamp)),
			Name: ev.Name,
		}
		BusEv.Data, err = skylab.FromJson(ev.Name, ev.Data)
		events = append(events, BusEv)
	}

	err = rows.Err()

	return events, err
}

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
