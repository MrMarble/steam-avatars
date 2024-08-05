package database

import (
	"database/sql"
	"os"

	_ "github.com/glebarez/go-sqlite"
)

type Database struct {
	db *sql.DB
}

func OpenDB() (*Database, error) {
	db, err := sql.Open("sqlite", "avatars.db?_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}

	init, err := os.ReadFile("database.sql")
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(string(init))
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

func (db *Database) Close() error {
	return db.db.Close()
}
