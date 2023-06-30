package gotelem

// this file implements the database functions to load/store/read from a sql database.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kschamplin/gotelem/skylab"
	sqlite3 "github.com/mattn/go-sqlite3"
)

func init() {
	sql.Register("custom_sqlite3", &sqlite3.SQLiteDriver{
		// TODO: add functions that convert between unix milliseconds and ISO 8601
	})
}

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
	// FIXME: only do this when it's a new database (instead warn the user about potential version mismatches)
	// TODO: store gotelem version (commit hash?) in DB (PRAGMA user_version)
	_, err = tdb.db.Exec(sqlDbUp)

	return tdb, err
}

// the sql commands to create the database.
const sqlDbUp = `
CREATE TABLE IF NOT EXISTS "bus_events" (
	"ts"	INTEGER NOT NULL, -- timestamp, unix milliseconds
	"id"	INTEGER NOT NULL, -- can ID
	"name"	TEXT NOT NULL, -- name of base packet
	"data"	TEXT NOT NULL CHECK(json_valid(data)) -- JSON object describing the data, including index if any
);

CREATE INDEX IF NOT EXISTS "ids_timestamped" ON "bus_events" (
	"id",
	"ts"	DESC
);

CREATE INDEX IF NOT EXISTS "times" ON "bus_events" (
	"ts" DESC
);

-- this table shows when we started/stopped logging.
CREATE TABLE "drive_records" (
	"id"	INTEGER NOT NULL UNIQUE, -- unique ID of the drive.
	"start_time"	INTEGER NOT NULL, -- when the drive started
	"end_time"	INTEGER, -- when it ended, or NULL if it's ongoing.
	"note"	TEXT, -- optional description of the segment/experiment/drive
	PRIMARY KEY("id" AUTOINCREMENT),
	CONSTRAINT "duration_valid" CHECK(end_time is null or start_time < end_time)
);

-- gps logs TODO: use GEOJSON/Spatialite tracks instead?
CREATE TABLE "position_logs" (
	"ts"	INTEGER NOT NULL,
	"source"	TEXT NOT NULL,
	"lat"	REAL NOT NULL,
	"lon"	REAL NOT NULL,
	"elevation"	REAL,
	CONSTRAINT "no_empty_source" CHECK(source is not "")
);

-- TODO: ensure only one "active" (end_time == null) drive record at a time using triggers/constraints/index
`

// sql sequence to tear down the database.
// not used often, but good to keep track of what's going on.
// Up() then Down() should result in an empty database.
const sqlDbDown = `
DROP TABLE "bus_events";
DROP INDEX "ids_timestamped";
DROP INDEX "times";

DROP TABLE "drive_records";
DROP TABLE "position_logs";
`

// sql expression to insert a bus event into the packets database.1
const sqlInsertEvent = `
INSERT INTO "bus_events" (time, can_id, name, packet) VALUES ($1, $2, $3, json($4));
`

// AddEvent adds the bus event to the database.
func (tdb *TelemDb) AddEventsCtx(ctx context.Context, events ...skylab.BusEvent) (err error) {
	//
	tx, err := tdb.db.BeginTx(ctx, nil)
	if err != nil {
		tx.Rollback()
		return
	}

	for _, b := range events {
		var j []byte
		j, err = json.Marshal(b.Data)

		if err != nil {
			tx.Rollback()
			return
		}
		tx.ExecContext(ctx, sqlInsertEvent, b.Timestamp.UnixMilli(), b.Id, b.Data.String(), j)
	}
	tx.Commit()
	return
}

func (tdb *TelemDb) AddEvents(events ...skylab.BusEvent) (err error) {

	return tdb.AddEventsCtx(context.Background(), events...)
}

/// Query fragment guide:
/// We need to be able to easily construct safe(!) and meaningful queries programatically
/// so we make some new types that can be turned into SQL fragments that go inside the where clause.
/// These all implement the QueryFrag interface, meaning the actual query function (that acts on the DB)
/// can deal with them agnostically. The Query function joins all the fragments it is given with AND.
/// to get OR,

// QueryFrag is anything that can be turned into a Query WHERE clause
type QueryFrag interface {
	Query() string
}

// QueryIdRange represents a range of IDs to select for, inclusive.
type QueryIdRange struct {
	Start uint32
	End   uint32
}

func (q *QueryIdRange) Query() string {
	return fmt.Sprintf("id BETWEEN %d AND %d", q.Start, q.End)
}

// QueryIds selects for individual CAN ids
type QueryIds []uint32

func (q QueryIds) Query() string {
	var idStrings []string
	for _, id := range q {
		idStrings = append(idStrings, strconv.FormatUint(uint64(id), 10))
	}
	return fmt.Sprintf("id IN (%s)", strings.Join(idStrings, ","))
}

// QueryTimeRange represents a query of a specific time range. For "before" or "after" queries,
// use time.Unix(0,0) or time.Now() in start and end respectively.
type QueryTimeRange struct {
	Start time.Time
	End   time.Time
}

func (q *QueryTimeRange) Query() string {
	startUnix := q.Start.UnixMilli()
	endUnix := q.End.UnixMilli()

	return fmt.Sprintf("ts BETWEEN %d AND %d", startUnix, endUnix)
}

type QueryNames []string

func (q QueryNames) Query() string {
	return fmt.Sprintf("name IN (%s)", strings.Join(q, ", "))
}

type QueryOr []QueryFrag

func (q QueryOr) Query() string {
	var qStrings []string
	for _, frag := range q {
		qStrings = append(qStrings, frag.Query())
	}
	return fmt.Sprintf("(%s)", strings.Join(qStrings, " OR "))
}

const eventQueryFmtString = `SELECT * FROM "bus_events" WHERE %s LIMIT %d`

// GetEvents is the mechanism to request underlying event data.
// it takes functions (which are defined in db.go) that modify the query,
// and then return the results.
func (tdb *TelemDb) GetEvents(limit int, where ...QueryFrag) (events []skylab.BusEvent, err error) {
	// Simple mechanism for combining query frags:
	// join with " AND ". To join expressions with or, use QueryOr
	var fragStr []string
	for _, f := range where {
		fragStr = append(fragStr, f.Query())
	}
	qString := fmt.Sprintf("SELECT * FROM \"bus_events\" WHERE %s LIMIT %d", strings.Join(fragStr, " AND "), limit)
	rows, err := tdb.db.Queryx(qString)
	if err != nil {
		return
	}
	defer rows.Close()

	if limit < 0 { //  special case: limit negative means unrestricted.
		events = make([]skylab.BusEvent, 0, 20)
	} else {
		events = make([]skylab.BusEvent, 0, limit)
	}
	// scan rows into busevent list...
	for rows.Next() {
		var ev skylab.RawJsonEvent
		err = rows.StructScan(&ev)
		if err != nil {
			return
		}

		BusEv := skylab.BusEvent{
			Timestamp: time.UnixMilli(int64(ev.Timestamp)),
			Id:        ev.Id,
		}
		BusEv.Data, err = skylab.FromJson(ev.Id, ev.Data)

		// FIXME: this is slow!
		events = append(events, BusEv)

	}

	err = rows.Err()

	return
}
