package database

import (
	"database/sql"

	_ "github.com/glebarez/go-sqlite"
)

type Database struct {
	db *sql.DB
}

func OpenDB() (*Database, error) {
	db, err := sql.Open("sqlite", "users.db?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}
