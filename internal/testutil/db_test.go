package testutil

import (
	"context"
	"testing"

	"familyshare/internal/db/sqlc"
)

func TestSetupTestDB(t *testing.T) {
	db, q, cleanup := SetupTestDB(t)
	defer cleanup()

	// Verify we can query tables
	ctx := context.Background()

	// Test albums table
	albums, err := q.ListAlbums(ctx, sqlc.ListAlbumsParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("failed to query albums: %v", err)
	}
	if len(albums) != 0 {
		t.Errorf("expected 0 albums in fresh DB, got %d", len(albums))
	}

	// Verify tables exist by checking schema
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('albums', 'photos', 'share_links', 'share_link_views', 'sessions', 'activity_events')").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query tables: %v", err)
	}
	if count != 6 {
		t.Errorf("expected 6 tables, got %d", count)
	}
}
