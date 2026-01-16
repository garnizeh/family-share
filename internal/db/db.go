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
