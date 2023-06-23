package gotelem

// this file implements the database functions to load/store/read from a sql database.
import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kschamplin/gotelem/skylab"
	_ "github.com/mattn/go-sqlite3"
)

type TelemDb struct {
	db *sqlx.DB
}

type TelemDbOption func(*TelemDb) error

func OpenTelemDb(path string, options ...TelemDbOption) (tdb *TelemDb, err error) {
	tdb = &TelemDb{}
	tdb.db, err = sqlx.Connect("sqlite3", path)
	if err != nil {
		return
	}
	// TODO: add options support.

	for _, fn := range options {
		err = fn(tdb)
		if err != nil {
			return
		}
	}

	// execute database up statement (better hope it is idempotent!)
	_, err = tdb.db.Exec(sqlDbUp)

	if err != nil {

	}

	return tdb, nil
}

// the sql commands to create the database.
const sqlDbUp = `
CREATE TABLE IF NOT EXISTS "bus_events" (
	"ts"	REAL NOT NULL, -- timestamp
	"id"	INTEGER NOT NULL, -- can ID
	"name"	TEXT NOT NULL, -- name of base packet
	"packet"	TEXT NOT NULL CHECK(json_valid(packet)) -- JSON object describing the data
);

CREATE INDEX IF NOT EXISTS "ids_timestamped" ON "bus_events" (
	"id",
	"ts"	DESC
);

CREATE INDEX IF NOT EXISTS "times" ON "bus_events" (
	"ts" DESC
);

`

// sql sequence to tear down the database.
// not used often, but good to keep track of what's going on.
// Up() then Down() should result in an empty database.
const sqlDbDown = `
DROP TABLE "bus_events";
DROP INDEX "ids_timestamped";
DROP INDEX "times";
`

// sql expression to insert a bus event into the packets database.1
const sqlInsertEvent = `
INSERT INTO "bus_events" (time, can_id, name, packet) VALUES ($1, $2, $3, json($4));
`

// AddEvent adds the bus event to the database.
func (tdb *TelemDb) AddEvents(events ...skylab.BusEvent) {
	//
	tx, err := tdb.db.Begin()
	if err != nil {
		tx.Rollback()
		return
	}

	for _, b := range events {
		j, err := json.Marshal(b.Data)

		if err != nil {
			tx.Rollback()
			return
		}
		tx.Exec(sqlInsertEvent, b.Timestamp, b.Id, b.Name, j)
	}
	tx.Commit()
}

// QueryIdString is a string that filters ids from the set. use ID query functions to
// create them.
type QueryIdString string

// QueryIds constructs a CAN Id filter for one or more distinct Ids.
// For a range of ids, use QueryIdRange(start, stop uint32)
func QueryIds(ids ...uint32) QueryIdString {
	// FIXME: zero elements case?
	var idsString []string
	for _, id := range ids {
		idsString = append(idsString, strconv.FormatUint(uint64(id), 10))
	}

	return QueryIdString("id IN (" + strings.Join(idsString, ",") + ")")
}

func QueryIdsInv(ids ...uint32) QueryIdString {

}

// QueryIdRange selects all IDs between start and end, *inclusive*.
// This function is preferred over a generated list of IDs.
func QueryIdRange(start, end uint32) QueryIdString {
	startString := strconv.FormatUint(uint64(start), 10)
	endString := strconv.FormatUint(uint64(end), 10)
	return QueryIdString("id BETWEEN " + startString + " AND " + endString)
}

// QueryIdRangeInv removes all IDs between start and end from the results.
// See QueryIdRange for more details.
func QueryIdRangeInv(start, end uint32) QueryIdString {
	return QueryIdString("NOT ") + QueryIdRange(start, end)
}

type QueryTimestampString string

// QueryDuration takes a start and end time and filters where the packets are between that time range.
func QueryDuration(start, end time.Time) QueryTimestampString {

	// the time in the database is a float, we have a time.Time so use unixNano() / 1e9 to float it.
	startString := strconv.FormatFloat(float64(start.UnixNano())/1e9, 'f', -1, 64)
	endString := strconv.FormatFloat(float64(start.UnixNano())/1e9, 'f', -1, 64)
	return QueryTimestampString("ts BETWEEN " + startString + " AND " + endString)
}

type QueryNameString string

func QueryNames(names ...string) QueryNameString

func QueryNamesInv(names ...string) QueryNameString

// Describes the parameters for an event query
type EventsQuery struct {
	Ids []QueryIdString // Ids contains a list of CAN ID filters that are OR'd together.

	Times []QueryTimestampString

	Names []QueryNameString

	Limit uint //  asdf
}

// GetEvents is the mechanism to request underlying event data.
// it takes functions (which are defined in db.go) that modify the query,
// and then return the results.
func (tdb *TelemDb) GetEvents(q *EventsQuery) []skylab.BusEvent {
	// if function is inverse, AND and OR are switched.
	// Demorgan's
	// how to know if function is inverted???
	return nil
}
