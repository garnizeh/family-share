package pipeline

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
)

func TestSaveProcessedImage_Success(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("STORAGE_PATH", tmp)

	dbPath := filepath.Join(tmp, "save_test.db")
	d, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer d.Close()

	ctx := context.Background()
	q := sqlc.New(d)
	alb, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{Title: "save-album"})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	data := []byte("webpdata")
	photoID, path, err := SaveProcessedImage(ctx, d, alb.ID, bytes.NewReader(data), 100, 50, len(data), "webp")
	if err != nil {
		t.Fatalf("SaveProcessedImage failed: %v", err)
	}
	if photoID == 0 {
		t.Fatalf("expected non-zero photoID")
	}
	// verify file exists and contents match
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %s, stat error: %v", path, err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if !bytes.Equal(b, data) {
		t.Fatalf("saved file contents differ")
	}

	// verify DB record
	got, err := q.GetPhoto(ctx, photoID)
	if err != nil {
		t.Fatalf("get photo: %v", err)
	}
	if got.Width != 100 || got.Height != 50 {
		t.Fatalf("unexpected dimensions, got %d x %d", got.Width, got.Height)
	}
}

func TestSaveProcessedImage_WriteFailure_RollsBack(t *testing.T) {
	tmp := t.TempDir()
	// create a file where STORAGE_PATH should be a directory to cause write failures
	blocked := filepath.Join(tmp, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	os.Setenv("STORAGE_PATH", blocked)

	dbPath := filepath.Join(tmp, "save_fail.db")
	d, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer d.Close()

	ctx := context.Background()
	q := sqlc.New(d)
	alb, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{Title: "save-album"})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	data := []byte("webpdata")
	_, _, err = SaveProcessedImage(ctx, d, alb.ID, bytes.NewReader(data), 100, 50, len(data), "webp")
	if err == nil {
		t.Fatalf("expected error when storage path is blocked")
	}

	// ensure no photos for album
	list, err := q.ListPhotosByAlbum(ctx, sqlc.ListPhotosByAlbumParams{AlbumID: alb.ID, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("list photos: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected no photos after failed save, got %d", len(list))
	}
}
