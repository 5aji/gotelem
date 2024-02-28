package db

// this file implements the database functions to load/store/read from a sql database.

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/kschamplin/gotelem/skylab"
	_ "github.com/mattn/go-sqlite3"
)


type TelemDb struct {
	db *sqlx.DB
}

// TelemDbOption lets you customize the behavior of the sqlite database
type TelemDbOption func(*TelemDb) error


// this function is internal use. It actually opens the database, but uses
// a raw path string instead of formatting one like the exported functions.
func openRawDb(rawpath string, options ...TelemDbOption) (tdb *TelemDb, err error) {
	tdb = &TelemDb{}
	tdb.db, err = sqlx.Connect("sqlite3", rawpath)
	if err != nil {
		return
	}
	for _, fn := range options {
		err = fn(tdb)
		if err != nil {
			return
		}
	}

	// perform any database migrations
	version, err := tdb.GetVersion()
	if err != nil {
		return
	}
	// TODO: use logging instead of printf
	fmt.Printf("starting version %d\n", version)

	version, err = RunMigrations(tdb)
	fmt.Printf("ending version %d\n", version)

	return tdb, err
}


// this string is used to open the read-write db.
// the extra options improve performance significantly.
const rwDbPathFmt = "file:%s?_journal_mode=wal&mode=rwc&_txlock=immediate&_timeout=10000"

// OpenTelemDb opens a new telemetry database at the given path.
func OpenTelemDb(path string, options ...TelemDbOption) (*TelemDb, error) {
	dbStr := fmt.Sprintf(rwDbPathFmt, path)
	return openRawDb(dbStr, options...)
}

func (tdb *TelemDb) GetVersion() (int, error) {
	var version int
	err := tdb.db.Get(&version, "PRAGMA user_version")
	return version, err
}

func (tdb *TelemDb) SetVersion(version int) error {
	stmt := fmt.Sprintf("PRAGMA user_version = %d", version)
	_, err := tdb.db.Exec(stmt)
	return err
}

// sql expression to insert a bus event into the packets database.1
const sqlInsertEvent =`INSERT INTO "bus_events" (ts, name, data) VALUES `

// AddEvent adds the bus event to the database.
func (tdb *TelemDb) AddEventsCtx(ctx context.Context, events ...skylab.BusEvent) (n int64, err error) {
	// edge case - zero events.
	if len(events) == 0 {
		return 0, nil
	}
	n = 0
	tx, err := tdb.db.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return
	}

	sqlStmt := sqlInsertEvent
	const rowSql = "(?, ?, json(?))"
	inserts := make([]string, len(events))
	vals := []interface{}{}
	idx := 0 // we have to manually increment, because sometimes we don't insert.
	for _, b := range events {
		inserts[idx] = rowSql
		var j []byte
		j, err = json.Marshal(b.Data)

		if err != nil {
			// we had some error turning the packet into json.
			continue // we silently skip.
		}

		vals = append(vals, b.Timestamp.UnixMilli(), b.Data.String(), j)
		idx++
	}

	// construct the full statement now
	sqlStmt = sqlStmt + strings.Join(inserts[:idx], ",")
	stmt, err := tx.PrepareContext(ctx, sqlStmt)
	// defer stmt.Close()
	if err != nil {
		return
	}
	res, err := stmt.ExecContext(ctx, vals...)
	if err != nil {
		return
	}
	n, err = res.RowsAffected()

	tx.Commit()
	return
}

func (tdb *TelemDb) AddEvents(events ...skylab.BusEvent) (int64, error) {

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
	qString := fmt.Sprintf(`SELECT * FROM "bus_events" WHERE %s LIMIT %d`, strings.Join(fragStr, " AND "), limit)
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
			Name:      ev.Name,
		}
		BusEv.Data, err = skylab.FromJson(ev.Name, ev.Data)

		// FIXME: this is slow!
		events = append(events, BusEv)

	}

	err = rows.Err()

	return
}

// GetActiveDrive finds the non-null drive and returns it, if any.
func (tdb *TelemDb) GetActiveDrive() (res int, err error) {
	err = tdb.db.Get(&res, "SELECT id FROM drive_records WHERE end_time IS NULL LIMIT 1")
	return
}

func (tdb *TelemDb) NewDrive(start time.Time, note string) {

}

func (tdb *TelemDb) EndDrive() {

}

func (tdb *TelemDb) UpdateDrive(id int, note string) {

}
