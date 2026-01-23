package metrics

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
)

func setupTestDB(t *testing.T) (*sql.DB, *sqlc.Queries) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	queries := sqlc.New(database)
	return database, queries
}

func TestLogUpload(t *testing.T) {
	database, queries := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	logger := New(database)

	// Create test album
	album, err := queries.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create album: %v", err)
	}

	// Create test photo
	photo, err := queries.CreatePhoto(ctx, sqlc.CreatePhotoParams{
		AlbumID:   album.ID,
		Filename:  "test.webp",
		Width:     800,
		Height:    600,
		SizeBytes: 12345,
		Format:    "webp",
	})
	if err != nil {
		t.Fatalf("Failed to create photo: %v", err)
	}

	// Log upload event
	err = logger.LogUpload(ctx, album.ID, photo.ID)
	if err != nil {
		t.Errorf("LogUpload failed: %v", err)
	}

	// Verify event was logged
	events, err := queries.ListRecentActivity(ctx, sqlc.ListRecentActivityParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to list activity: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.EventType != "upload" {
		t.Errorf("Expected event type 'upload', got '%s'", event.EventType)
	}
	if !event.AlbumID.Valid || event.AlbumID.Int64 != album.ID {
		t.Errorf("Expected album_id %d, got %v", album.ID, event.AlbumID)
	}
	if !event.PhotoID.Valid || event.PhotoID.Int64 != photo.ID {
		t.Errorf("Expected photo_id %d, got %v", photo.ID, event.PhotoID)
	}
}

func TestLogShareView(t *testing.T) {
	database, queries := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	logger := New(database)

	// Create test album
	album, err := queries.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create album: %v", err)
	}

	// Create test share link
	link, err := queries.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "test-token",
		TargetType: "album",
		TargetID:   album.ID,
		ExpiresAt:  sql.NullTime{Time: time.Now().UTC().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create share link: %v", err)
	}

	// Log share view event
	err = logger.LogShareView(ctx, link.ID)
	if err != nil {
		t.Errorf("LogShareView failed: %v", err)
	}

	// Verify event was logged
	events, err := queries.ListRecentActivity(ctx, sqlc.ListRecentActivityParams{
		Limit:  10,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("Failed to list activity: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}

	event := events[0]
	if event.EventType != "share_view" {
		t.Errorf("Expected event type 'share_view', got '%s'", event.EventType)
	}
	if !event.ShareLinkID.Valid || event.ShareLinkID.Int64 != link.ID {
		t.Errorf("Expected share_link_id %d, got %v", link.ID, event.ShareLinkID)
	}
}

func TestGetStats(t *testing.T) {
	database, queries := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	logger := New(database)

	// Create test album
	album, err := queries.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create album: %v", err)
	}

	// Create events at different times
	now := time.Now().UTC()

	// Recent upload (within 7 days)
	err = queries.CreateActivityEvent(ctx, sqlc.CreateActivityEventParams{
		EventType:   "upload",
		AlbumID:     sql.NullInt64{Int64: album.ID, Valid: true},
		PhotoID:     sql.NullInt64{Int64: 1, Valid: true},
		ShareLinkID: sql.NullInt64{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create recent upload event: %v", err)
	}

	// Old upload (20 days ago) - manipulate created_at manually
	_, err = database.Exec(`
		INSERT INTO activity_events (event_type, album_id, photo_id, share_link_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, "upload", album.ID, 2, nil, now.Add(-20*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create old upload event: %v", err)
	}

	// Very old upload (60 days ago)
	_, err = database.Exec(`
		INSERT INTO activity_events (event_type, album_id, photo_id, share_link_id, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, "upload", album.ID, 3, nil, now.Add(-60*24*time.Hour))
	if err != nil {
		t.Fatalf("Failed to create very old upload event: %v", err)
	}

	// Recent share view (within 7 days)
	err = queries.CreateActivityEvent(ctx, sqlc.CreateActivityEventParams{
		EventType:   "share_view",
		AlbumID:     sql.NullInt64{Valid: false},
		PhotoID:     sql.NullInt64{Valid: false},
		ShareLinkID: sql.NullInt64{Int64: 1, Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create recent share view event: %v", err)
	}

	// Get stats
	stats, err := logger.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	// Verify 7-day stats (1 upload, 1 share view)
	if stats.Uploads7Days != 1 {
		t.Errorf("Expected 1 upload in last 7 days, got %d", stats.Uploads7Days)
	}
	if stats.ShareViews7Days != 1 {
		t.Errorf("Expected 1 share view in last 7 days, got %d", stats.ShareViews7Days)
	}

	// Verify 30-day stats (2 uploads, 1 share view)
	if stats.Uploads30Days != 2 {
		t.Errorf("Expected 2 uploads in last 30 days, got %d", stats.Uploads30Days)
	}
	if stats.ShareViews30Days != 1 {
		t.Errorf("Expected 1 share view in last 30 days, got %d", stats.ShareViews30Days)
	}
}

func TestGetStatsEmpty(t *testing.T) {
	database, _ := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()
	logger := New(database)

	// Get stats with no events
	stats, err := logger.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	// Verify all stats are zero
	if stats.Uploads7Days != 0 {
		t.Errorf("Expected 0 uploads in last 7 days, got %d", stats.Uploads7Days)
	}
	if stats.Uploads30Days != 0 {
		t.Errorf("Expected 0 uploads in last 30 days, got %d", stats.Uploads30Days)
	}
	if stats.AlbumViews7Days != 0 {
		t.Errorf("Expected 0 album views in last 7 days, got %d", stats.AlbumViews7Days)
	}
	if stats.AlbumViews30Days != 0 {
		t.Errorf("Expected 0 album views in last 30 days, got %d", stats.AlbumViews30Days)
	}
	if stats.ShareViews7Days != 0 {
		t.Errorf("Expected 0 share views in last 7 days, got %d", stats.ShareViews7Days)
	}
	if stats.ShareViews30Days != 0 {
		t.Errorf("Expected 0 share views in last 30 days, got %d", stats.ShareViews30Days)
	}
}

func TestEventTypeConstants(t *testing.T) {
	if EventUpload != "upload" {
		t.Errorf("Expected EventUpload to be 'upload', got '%s'", EventUpload)
	}
	if EventAlbumView != "album_view" {
		t.Errorf("Expected EventAlbumView to be 'album_view', got '%s'", EventAlbumView)
	}
	if EventPhotoView != "photo_view" {
		t.Errorf("Expected EventPhotoView to be 'photo_view', got '%s'", EventPhotoView)
	}
	if EventShareView != "share_view" {
		t.Errorf("Expected EventShareView to be 'share_view', got '%s'", EventShareView)
	}
}
