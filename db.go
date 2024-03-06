package gotelem

// this file implements the database functions to load/store/read from a sql database.

import (
	"context"
	"encoding/json"
	"errors"
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
func OpenRawDb(rawpath string, options ...TelemDbOption) (tdb *TelemDb, err error) {
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
const ProductionDbURI = "file:%s?_journal_mode=wal&mode=rwc&_txlock=immediate&_timeout=10000"

// OpenTelemDb opens a new telemetry database at the given path.
func OpenTelemDb(path string, options ...TelemDbOption) (*TelemDb, error) {
	dbStr := fmt.Sprintf(ProductionDbURI, path)
	return OpenRawDb(dbStr, options...)
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
const sqlInsertEvent = `INSERT INTO "bus_events" (ts, name, data) VALUES `

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

// QueryModifier augments SQL strings.
type QueryModifier interface {
	ModifyStatement(*strings.Builder) error
}

// LimitOffsetModifier is a modifier to support pagniation.
type LimitOffsetModifier struct {
	Limit  int
	Offset int
}

func (l *LimitOffsetModifier) ModifyStatement(sb *strings.Builder) error {
	clause := fmt.Sprintf(" LIMIT %d OFFSET %d", l.Limit, l.Offset)
	sb.WriteString(clause)
	return nil
}

// BusEventFilter is a filter for bus events.
type BusEventFilter struct {
	Names     []string  // The name(s) of packets to filter for
	StartTime time.Time // Starting time range. All packets >= StartTime
	EndTime   time.Time // Ending time range. All packets <= EndTime
	Indexes   []int     // The specific index of the packets to index.
}

// now we can optionally add a limit.

func (tdb *TelemDb) GetPackets(ctx context.Context, filter BusEventFilter, options ...QueryModifier) ([]skylab.BusEvent, error) {
	// construct a simple
	var whereFrags = make([]string, 0)

	// if we're filtering by names, add a where clause for it.
	if len(filter.Names) > 0 {
		// we have to quote our individual names
		names := strings.Join(filter.Names, `", "`)
		qString := fmt.Sprintf(`name IN ("%s")`, names)
		whereFrags = append(whereFrags, qString)
	}
	// TODO: identify if we need a special case for both time ranges
	// using BETWEEN since apparenlty that can be better?

	// next, check if we have a start/end time, add constraints
	if !filter.EndTime.IsZero() {
		qString := fmt.Sprintf("ts <= %d", filter.EndTime.UnixMilli())
		whereFrags = append(whereFrags, qString)
	}
	if !filter.StartTime.IsZero() {
		// we have an end range
		qString := fmt.Sprintf("ts >= %d", filter.StartTime.UnixMilli())
		whereFrags = append(whereFrags, qString)
	}
	if len(filter.Indexes) > 0 {
		s := make([]string, 0)
		for _, idx := range filter.Indexes {
			s = append(s, fmt.Sprint(idx))
		}
		idxs := strings.Join(s, ", ")
		qString := fmt.Sprintf(`idx in (%s)`, idxs)
		whereFrags = append(whereFrags, qString)
	}

	sb := strings.Builder{}
	sb.WriteString(`SELECT ts, name, data from "bus_events"`)
	// construct the full statement.
	if len(whereFrags) > 0 {
		// use the where clauses.
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(whereFrags, " AND "))
	}

	sb.WriteString(" ORDER BY ts DESC")

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

		BusEv := skylab.BusEvent{
			Timestamp: time.UnixMilli(int64(ev.Timestamp)),
			Name:      ev.Name,
		}
		BusEv.Data, err = skylab.FromJson(ev.Name, ev.Data)
		if err != nil {
			return events, nil
		}
		events = append(events, BusEv)
	}

	err = rows.Err()

	return events, err
}

// We now need a different use-case: we would like to extract a value from
// a specific packet.

// Datum is a single measurement - it is more granular than a packet.
// the classic example is bms_measurement.current
type Datum struct {
	Timestamp time.Time `db:"timestamp" json:"ts"`
	Value     any       `db:"val" json:"val"`
}

// GetValues queries the database for values in a given time range.
// A value is a specific data point. For example, bms_measurement.current
// would be a value.
func (tdb *TelemDb) GetValues(ctx context.Context, filter BusEventFilter,
	field string, opts ...QueryModifier) ([]Datum, error) {
	// this fragment uses json_extract from sqlite to get a single
	// nested value.
	sb := strings.Builder{}
	sb.WriteString(`SELECT ts as timestamp, json_extract(data, '$.' || ?) as val FROM bus_events WHERE `)
	if len(filter.Names) != 1 {
		return nil, errors.New("invalid number of names")
	}
	whereFrags := []string{"name is ?"}

	if !filter.StartTime.IsZero() {
		qString := fmt.Sprintf("ts >= %d", filter.StartTime.UnixMilli())
		whereFrags = append(whereFrags, qString)
	}

	if !filter.EndTime.IsZero() {
		qString := fmt.Sprintf("ts <= %d", filter.EndTime.UnixMilli())
		whereFrags = append(whereFrags, qString)
	}
	if len(filter.Indexes) > 0 {
		s := make([]string, 0)
		for _, idx := range filter.Indexes {
			s = append(s, fmt.Sprint(idx))
		}
		idxs := strings.Join(s, ", ")
		qString := fmt.Sprintf(`idx in (%s)`, idxs)
		whereFrags = append(whereFrags, qString)
	}
	// join qstrings with AND
	sb.WriteString(strings.Join(whereFrags, " AND "))

	sb.WriteString(" ORDER BY ts DESC")

	for _, m := range opts {
		if m == nil {
			continue
		}
		m.ModifyStatement(&sb)
	}
	rows, err := tdb.db.QueryxContext(ctx, sb.String(), field, filter.Names[0])
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	data := make([]Datum, 0, 10)
	for rows.Next() {
		var d Datum = Datum{}
		var ts int64
		err = rows.Scan(&ts, &d.Value)
		d.Timestamp = time.UnixMilli(ts)

		if err != nil {
			fmt.Print(err)
			return data, err
		}
		data = append(data, d)
	}
	fmt.Print(rows.Err())

	return data, nil
}

// AddDocument inserts a new document to the store if it is unique and valid.
func (tdb *TelemDb) AddDocument(ctx context.Context, obj json.RawMessage) error {
	const insertStmt = `INSERT INTO openmct_objects (data) VALUES (json(?))`
	_, err := tdb.db.ExecContext(ctx, insertStmt, obj)
	return err
}

// DocumentNotFoundError is when the underlying document cannot be found.
type DocumentNotFoundError string

func (e DocumentNotFoundError) Error() string {
	return fmt.Sprintf("document could not find key: %s", string(e))
}


// UpdateDocument replaces the entire contents of a document matching
// the given key. Note that the key is derived from the document,
// and no checks are done to ensure that the new key is the same.
func (tdb *TelemDb) UpdateDocument(ctx context.Context, key string,
	obj json.RawMessage) error {

	const upd = `UPDATE openmct_objects SET data = json(?) WHERE key IS ?`
	r, err := tdb.db.ExecContext(ctx, upd, obj, key)
	if err != nil {
		return err
	}
	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return DocumentNotFoundError(key)
	}
	return err
}


// GetDocument gets the document matching the corresponding key.
func (tdb *TelemDb) GetDocument(ctx context.Context, key string) (json.RawMessage, error) {
	const get = `SELECT data FROM openmct_objects WHERE key IS ?`

	row := tdb.db.QueryRowxContext(ctx, get, key)

	var res []byte // VERY important, json.RawMessage won't work here
	// since the scan function does not look at underlying types.
	row.Scan(&res)

	if len(res) == 0 {
		return nil, DocumentNotFoundError(key)
	}

	return res, nil
}

// GetAllDocuments returns all documents in the database.
func (tdb *TelemDb) GetAllDocuments(ctx context.Context) ([]json.RawMessage, error) {
	const getall = `SELECT data FROM openmct_objects`;

	rows, err := tdb.db.QueryxContext(ctx, getall)
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	docs := make([]json.RawMessage, 0)
	for rows.Next() {
		var j json.RawMessage
		rows.Scan(&j)
		docs = append(docs, j)
	}
	return docs, nil
}

// DeleteDocument removes a document from the store, or errors
// if it does not exist.
func (tdb *TelemDb) DeleteDocument(ctx context.Context, key string) error {
	const del = `DELETE FROM openmct_objects WHERE key IS ?`
	res, err := tdb.db.ExecContext(ctx, del, key)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return DocumentNotFoundError(key)
	}
	return err
}
