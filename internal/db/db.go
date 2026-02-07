package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"

	fssql "familyshare/sql"
)

// InitDB opens a SQLite database at path and applies embedded migrations.
func InitDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode (better concurrency)
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable wal: %w", err)
	}

	// Set busy timeout (e.g., 5000ms = 5s)
	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := ApplyMigrations(db, fssql.MigrationsFS); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
