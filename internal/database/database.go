package database

import (
	"time"

	_ "github.com/glebarez/go-sqlite"
	"github.com/valkey-io/valkey-go"
)

type Database struct {
	client valkey.Client
}

func OpenDB(endpoint string) (*Database, error) {
	db, err := valkey.NewClient(valkey.ClientOption{
		InitAddress:      []string{endpoint},
		ConnWriteTimeout: 3 * time.Second,
		DisableRetry:     true,
	})
	if err != nil {
		return nil, err
	}

	return &Database{db}, nil
}

func (db *Database) Close() {
	db.client.Close()
}
