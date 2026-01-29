package janitor

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/storage"
)

func setupTestDB(t *testing.T) (*sql.DB, *sqlc.Queries, string) {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	queries := sqlc.New(database)
	return database, queries, tmpDir
}

func TestJanitorDeleteExpiredSessions(t *testing.T) {
	database, queries, tmpDir := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create an active session
	activeSession, err := queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        "active-session",
		UserID:    "admin",
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Failed to create active session: %v", err)
	}

	// Create an expired session
	expiredSession, err := queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        "expired-session",
		UserID:    "admin",
		ExpiresAt: time.Now().UTC().Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Create janitor and run cleanup
	j := New(Config{
		DB:          database,
		StoragePath: tmpDir,
		Interval:    1 * time.Hour,
	})

	j.deleteExpiredSessions(ctx)

	// Verify active session still exists
	_, err = queries.GetSession(ctx, activeSession.ID)
	if err != nil {
		t.Errorf("Active session should still exist, got error: %v", err)
	}

	// Verify expired session was deleted
	_, err = queries.GetSession(ctx, expiredSession.ID)
	if err == nil {
		t.Error("Expired session should have been deleted")
	}
}

func TestJanitorDeleteExpiredShareLinks(t *testing.T) {
	database, queries, tmpDir := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create an album for testing
	album, err := queries.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create album: %v", err)
	}

	// Create an active share link
	activeLink, err := queries.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "active-token",
		TargetType: "album",
		TargetID:   album.ID,
		ExpiresAt:  sql.NullTime{Time: time.Now().UTC().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create active share link: %v", err)
	}

	// Create an expired share link
	expiredLink, err := queries.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "expired-token",
		TargetType: "album",
		TargetID:   album.ID,
		ExpiresAt:  sql.NullTime{Time: time.Now().UTC().Add(-24 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create expired share link: %v", err)
	}

	// Create a revoked share link
	revokedLink, err := queries.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "revoked-token",
		TargetType: "album",
		TargetID:   album.ID,
		ExpiresAt:  sql.NullTime{Time: time.Now().UTC().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create revoked share link: %v", err)
	}

	// Revoke the link
	err = queries.RevokeShareLink(ctx, revokedLink.ID)
	if err != nil {
		t.Fatalf("Failed to revoke share link: %v", err)
	}

	// Create janitor and run cleanup
	j := New(Config{
		DB:          database,
		StoragePath: tmpDir,
		Interval:    1 * time.Hour,
	})

	j.deleteExpiredShareLinks(ctx)

	// Verify active link still exists
	_, err = queries.GetShareLink(ctx, activeLink.ID)
	if err != nil {
		t.Errorf("Active share link should still exist, got error: %v", err)
	}

	// Verify expired link was deleted
	_, err = queries.GetShareLink(ctx, expiredLink.ID)
	if err == nil {
		t.Error("Expired share link should have been deleted")
	}

	// Verify revoked link was deleted
	_, err = queries.GetShareLink(ctx, revokedLink.ID)
	if err == nil {
		t.Error("Revoked share link should have been deleted")
	}
}

func TestJanitorDeleteOrphanedPhotos(t *testing.T) {
	database, queries, tmpDir := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create an album
	album, err := queries.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create album: %v", err)
	}

	// Create a photo in the album
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

	// Create photo file on disk
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	photoPath := storage.PhotoPathAt(tmpDir, photo.AlbumID, photo.ID, photo.Format, createdAt)
	if err := os.MkdirAll(filepath.Dir(photoPath), 0755); err != nil {
		t.Fatalf("Failed to create photo directory: %v", err)
	}
	if err := os.WriteFile(photoPath, []byte("test photo"), 0644); err != nil {
		t.Fatalf("Failed to create photo file: %v", err)
	}

	// Create thumbnail file
	thumbPath := storage.ThumbnailPathAt(tmpDir, photo.AlbumID, photo.ID, createdAt)
	if err := os.WriteFile(thumbPath, []byte("test thumb"), 0644); err != nil {
		t.Fatalf("Failed to create thumbnail file: %v", err)
	}

	// Manually delete the photo from DB to simulate orphaned state
	// (bypassing CASCADE to test janitor's orphan detection)
	// First, update any references to avoid constraint violations
	_, err = database.Exec("UPDATE albums SET cover_photo_id = NULL WHERE cover_photo_id = ?", photo.ID)
	if err != nil {
		t.Fatalf("Failed to clear cover photo: %v", err)
	}

	// Now delete the album, which will CASCADE delete the photo
	// But we'll track the photo info before deletion for file cleanup
	photoForCleanup := photo // Save photo info before CASCADE deletes it

	err = queries.DeleteAlbum(ctx, album.ID)
	if err != nil {
		t.Fatalf("Failed to delete album: %v", err)
	}

	// Verify files exist before cleanup
	if _, err := os.Stat(photoPath); os.IsNotExist(err) {
		t.Fatal("Photo file should exist before cleanup")
	}
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		t.Fatal("Thumbnail file should exist before cleanup")
	}

	// Manually delete files to simulate what janitor should do
	// Since CASCADE already deleted from DB, we test file cleanup directly
	if err := os.Remove(photoPath); err != nil {
		t.Fatalf("Test cleanup: failed to remove photo: %v", err)
	}
	if err := os.Remove(thumbPath); err != nil && !os.IsNotExist(err) {
		t.Fatalf("Test cleanup: failed to remove thumb: %v", err)
	}

	// Temporarily disable foreign keys to insert orphaned photo
	_, err = database.Exec("PRAGMA foreign_keys = OFF")
	if err != nil {
		t.Fatalf("Failed to disable foreign keys: %v", err)
	}

	// Store photo back in DB with non-existent album_id to create true orphan
	_, err = database.Exec(
		"INSERT INTO photos (id, album_id, filename, width, height, size_bytes, format, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		photoForCleanup.ID+1000, 99999, photoForCleanup.Filename, photoForCleanup.Width, photoForCleanup.Height, photoForCleanup.SizeBytes, photoForCleanup.Format, photoForCleanup.CreatedAt,
	)
	if err != nil {
		t.Fatalf("Failed to create orphan photo: %v", err)
	}

	// Re-enable foreign keys
	_, err = database.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to re-enable foreign keys: %v", err)
	}

	// Create file for this orphaned photo
	orphanCreatedAt := time.Now().UTC()
	if photoForCleanup.CreatedAt.Valid {
		orphanCreatedAt = photoForCleanup.CreatedAt.Time.UTC()
	}
	orphanPhotoPath := storage.PhotoPathAt(tmpDir, 99999, photoForCleanup.ID+1000, photoForCleanup.Format, orphanCreatedAt)
	if err := os.MkdirAll(filepath.Dir(orphanPhotoPath), 0755); err != nil {
		t.Fatalf("Failed to create orphan photo directory: %v", err)
	}
	if err := os.WriteFile(orphanPhotoPath, []byte("orphan photo"), 0644); err != nil {
		t.Fatalf("Failed to create orphan photo file: %v", err)
	}
	orphanThumbPath := storage.ThumbnailPathAt(tmpDir, 99999, photoForCleanup.ID+1000, orphanCreatedAt)
	if err := os.WriteFile(orphanThumbPath, []byte("orphan thumb"), 0644); err != nil {
		t.Fatalf("Failed to create orphan thumbnail file: %v", err)
	}

	// Create janitor and run cleanup
	j := New(Config{
		DB:          database,
		StoragePath: tmpDir,
		Interval:    1 * time.Hour,
	})

	j.deleteOrphanedPhotos(ctx)

	// Verify orphaned photo was deleted from database
	_, err = queries.GetPhoto(ctx, photoForCleanup.ID+1000)
	if err == nil {
		t.Error("Orphaned photo should have been deleted from database")
	}

	// Verify orphaned photo files were deleted from disk
	if _, err := os.Stat(orphanPhotoPath); !os.IsNotExist(err) {
		t.Error("Orphaned photo file should have been deleted")
	}
	if _, err := os.Stat(orphanThumbPath); !os.IsNotExist(err) {
		t.Error("Orphaned thumbnail file should have been deleted")
	}
}

func TestJanitorGracefulShutdown(t *testing.T) {
	database, _, tmpDir := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	j := New(Config{
		DB:          database,
		StoragePath: tmpDir,
		Interval:    100 * time.Millisecond,
	})

	// Start janitor
	j.Start(ctx)

	// Let it run for a bit
	time.Sleep(150 * time.Millisecond)

	// Stop janitor
	done := make(chan struct{})
	go func() {
		j.Stop()
		close(done)
	}()

	// Should stop within reasonable time
	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Error("Janitor did not stop gracefully within timeout")
	}
}

func TestJanitorDefaultInterval(t *testing.T) {
	database, _, tmpDir := setupTestDB(t)
	defer database.Close()

	j := New(Config{
		DB:          database,
		StoragePath: tmpDir,
		// No interval set - should default to 6 hours
	})

	expected := 6 * time.Hour
	if j.interval != expected {
		t.Errorf("Expected default interval of %v, got %v", expected, j.interval)
	}
}

func TestJanitorRunsOnStartup(t *testing.T) {
	database, queries, tmpDir := setupTestDB(t)
	defer database.Close()

	ctx := context.Background()

	// Create an expired session
	_, err := queries.CreateSession(ctx, sqlc.CreateSessionParams{
		ID:        "expired-session",
		UserID:    "admin",
		ExpiresAt: time.Now().UTC().Add(-24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Failed to create expired session: %v", err)
	}

	// Create janitor with very long interval
	j := New(Config{
		DB:          database,
		StoragePath: tmpDir,
		Interval:    24 * time.Hour, // Won't run from timer
	})

	// Start janitor
	j.Start(ctx)

	// Give it a moment to run initial cleanup
	time.Sleep(100 * time.Millisecond)

	// Stop immediately
	j.Stop()

	// Verify cleanup ran on startup (expired session should be gone)
	_, err = queries.GetSession(ctx, "expired-session")
	if err == nil {
		t.Error("Janitor should have run cleanup on startup and deleted expired session")
	}
}

func TestJanitorCleanupTempFiles_CustomDir(t *testing.T) {
	database, _, tmpDir := setupTestDB(t)
	defer database.Close()

	customTemp := filepath.Join(tmpDir, "tmp_uploads")
	if err := os.MkdirAll(customTemp, 0o700); err != nil {
		t.Fatalf("mkdir custom temp: %v", err)
	}

	oldFile := filepath.Join(customTemp, "upload-old.tmp")
	if err := os.WriteFile(oldFile, []byte("x"), 0o600); err != nil {
		t.Fatalf("write old temp file: %v", err)
	}
	oldTime := time.Now().UTC().Add(-20 * time.Minute)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes old temp file: %v", err)
	}

	newFile := filepath.Join(customTemp, "upload-new.tmp")
	if err := os.WriteFile(newFile, []byte("y"), 0o600); err != nil {
		t.Fatalf("write new temp file: %v", err)
	}

	j := New(Config{
		DB:            database,
		StoragePath:   tmpDir,
		TempUploadDir: customTemp,
		Interval:      1 * time.Hour,
	})

	j.cleanupTempFiles()

	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Fatalf("expected old temp file to be removed")
	}
	if _, err := os.Stat(newFile); err != nil {
		t.Fatalf("expected new temp file to remain, stat error: %v", err)
	}
}
