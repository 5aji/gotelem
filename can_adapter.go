package gotelem

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// this file implements a CAN adapter for the sqlite db.

type CanDB struct {
	Db *sqlx.DB
}

func (cdb *CanDB) Send(_ *Frame) error {
	panic("not implemented") // TODO: Implement
}

func (cdb *CanDB) Recv() (*Frame, error) {
	panic("not implemented") // TODO: Implement
}

func NewCanDB()
