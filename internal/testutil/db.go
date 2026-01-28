package testutil

import (
	"database/sql"
	"testing"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	fssql "familyshare/sql"

	_ "modernc.org/sqlite"
)

// SetupTestDB creates a temporary in-memory SQLite database with migrations applied.
// Returns the database connection and a cleanup function that should be deferred.
func SetupTestDB(t *testing.T) (*sql.DB, *sqlc.Queries, func()) {
	t.Helper()

	database, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	// Ensure in-memory DB uses a single connection to avoid per-connection isolation
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)

	// Enable foreign keys
	if _, err := database.Exec("PRAGMA foreign_keys = ON"); err != nil {
		database.Close()
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Apply migrations
	if err := db.ApplyMigrations(database, fssql.MigrationsFS); err != nil {
		database.Close()
		t.Fatalf("failed to apply migrations: %v", err)
	}

	// Verify critical tables exist
	var count int
	err = database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('albums', 'photos', 'share_links', 'share_link_views', 'sessions')").Scan(&count)
	if err != nil {
		database.Close()
		t.Fatalf("failed to verify tables: %v", err)
	}
	if count != 5 {
		database.Close()
		t.Fatalf("expected 5 critical tables, found %d", count)
	}

	queries := sqlc.New(database)

	cleanup := func() {
		database.Close()
	}

	return database, queries, cleanup
}
