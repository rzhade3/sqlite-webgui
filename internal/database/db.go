package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn     *sql.DB
	path     string
	readonly bool
}

func New(dbPath string, readonly bool) (*DB, error) {
	connStr := dbPath
	if readonly {
		connStr = fmt.Sprintf("file:%s?mode=ro", dbPath)
	}

	conn, err := sql.Open("sqlite", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &DB{
		conn:     conn,
		path:     dbPath,
		readonly: readonly,
	}, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

func (db *DB) IsReadOnly() bool {
	return db.readonly
}
