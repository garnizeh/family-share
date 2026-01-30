package pipeline

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/storage"

	"github.com/gen2brain/avif"
)

func makeJPEG(t *testing.T, w, h int) *bytes.Reader {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// fill with a color
	for y := 0; y < h; y++ {
		for x := range w {
			img.Set(x, y, color.RGBA{R: 100, G: 150, B: 200, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return bytes.NewReader(buf.Bytes())
}

func TestProcessAndSave_Success(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("STORAGE_PATH", tmp)

	dbPath := filepath.Join(tmp, "test.db")
	d, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer d.Close()

	ctx := WithSkipUploadEvent(context.Background())
	q := sqlc.New(d)
	alb, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{Title: "test"})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	r := makeJPEG(t, 200, 100)
	photo, err := ProcessAndSave(ctx, d, alb.ID, r, 10<<20, tmp)
	if err != nil {
		t.Fatalf("process and save failed: %v", err)
	}

	// verify file exists
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	path := storage.PhotoPathAt(tmp, alb.ID, photo.ID, photo.Format, createdAt)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %s, stat error: %v", path, err)
	}
	if photo.SizeBytes <= 0 {
		t.Fatalf("expected photo size recorded, got %d", photo.SizeBytes)
	}
}

func TestProcessAndSave_WriteFailure_RollsBack(t *testing.T) {
	tmp := t.TempDir()
	// create a file at the base storage path to block directory creation
	blocked := filepath.Join(tmp, "blocked")
	if err := os.WriteFile(blocked, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker: %v", err)
	}
	os.Setenv("STORAGE_PATH", blocked)

	dbPath := filepath.Join(tmp, "test2.db")
	d, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer d.Close()

	ctx := WithSkipUploadEvent(context.Background())
	q := sqlc.New(d)
	alb, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{Title: "test"})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	r := makeJPEG(t, 200, 100)
	_, err = ProcessAndSave(ctx, d, alb.ID, r, 10<<20, blocked)
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

func TestProcessAndSaveWithFormat_AVIF(t *testing.T) {
	tmp := t.TempDir()
	os.Setenv("STORAGE_PATH", tmp)

	dbPath := filepath.Join(tmp, "test-avif.db")
	d, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	defer d.Close()

	ctx := WithSkipUploadEvent(context.Background())
	q := sqlc.New(d)
	alb, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{Title: "test avif"})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	r := makeJPEG(t, 200, 100)
	photo, err := ProcessAndSaveWithFormat(ctx, d, alb.ID, r, 10<<20, tmp, "avif")
	if err != nil {
		t.Fatalf("process and save failed: %v", err)
	}
	if photo.Format != "avif" {
		t.Fatalf("expected format avif, got %s", photo.Format)
	}

	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	path := storage.PhotoPathAt(tmp, alb.ID, photo.ID, photo.Format, createdAt)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read avif file: %v", err)
	}
	if _, err := avif.Decode(bytes.NewReader(data)); err != nil {
		t.Fatalf("decode avif failed: %v", err)
	}
}
