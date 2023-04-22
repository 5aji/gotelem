package db

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/kschamplin/gotelem/internal/can"
)

// this file implements a CAN adapter for the sqlite db.

type CanDB struct {
	Db *sqlx.DB
}

func (cdb *CanDB) Send(_ *can.Frame) error {
	panic("not implemented") // TODO: Implement
}

func (cdb *CanDB) Recv() (*can.Frame, error) {
	panic("not implemented") // TODO: Implement
}
