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

func (tdb *TelemDb) GetPackets(ctx context.Context, filter BusEventFilter, options ...QueryModifier) ([]skylab.BusEvent, error) {
	// construct a simple
	var whereFrags = make([]string, 0)

	// if we're filtering by names, add a where clause for it.
	if len(filter.Names) > 0 {
		names := strings.Join(filter.Names, ", ")
		qString := fmt.Sprintf(`name IN ("%s")`, names)
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
	sb.WriteString(`SELECT * from "bus_events"`)
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
func (tdb *TelemDb) GetValues(ctx context.Context, bef BusEventFilter,
	field string, opts ...QueryModifier) ([]Datum, error) {
	// this fragment uses json_extract from sqlite to get a single
	// nested value.
	sb := strings.Builder{}
	sb.WriteString(`SELECT ts as timestamp, json_extract(data, '$.' || ?) as val FROM bus_events WHERE `)
	if len(bef.Names) != 1 {
		return nil, errors.New("invalid number of names")
	}

	qStrings := []string{"name is ?"}
	// add timestamp limit.
	if !bef.TimerangeStart.IsZero() {
		qString := fmt.Sprintf("ts >= %d", bef.TimerangeStart.UnixMilli())
		qStrings = append(qStrings, qString)
	}

	if !bef.TimerangeEnd.IsZero() {
		qString := fmt.Sprintf("ts <= %d", bef.TimerangeEnd.UnixMilli())
		qStrings = append(qStrings, qString)
	}
	// join qstrings with AND
	sb.WriteString(strings.Join(qStrings, " AND "))

	for _, m := range opts {
		if m == nil {
			continue
		}
		m.ModifyStatement(&sb)
	}
	rows, err := tdb.db.QueryxContext(ctx, sb.String(), field, bef.Names[0])
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

// PacketDef is a database packet model
type PacketDef struct {
	Name        string
	Description string
	Id          int
}

type FieldDef struct {
	Name    string
	SubName string
	Packet  string
	Type    string
}

// PacketNotFoundError is when a matching packet cannot be found.
type PacketNotFoundError string

func (e *PacketNotFoundError) Error() string {
	return "packet not found: " + string(*e)
}

// GetPacketDefN retrieves a packet matching the given name, if it exists.
// returns PacketNotFoundError if a matching packet could not be found.
func (tdb *TelemDb) GetPacketDefN(name string) (*PacketDef, error) {
	return nil, nil
}

// GetPacketDefF retrieves the parent packet for a given field.
// This function cannot return PacketNotFoundError since we have SQL FKs enforcing.
func (tdb *TelemDb) GetPacketDefF(field FieldDef) (*PacketDef, error) {
	return nil, nil
}

// GetFieldDefs returns the given fields for a given packet definition.
func (tdb *TelemDb) GetFieldDefs(pkt PacketDef) ([]FieldDef, error) {
	return nil, nil
}
