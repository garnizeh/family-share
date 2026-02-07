package handler_test

import (
	"context"
	"database/sql"
	"image"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/chai2010/webp"
	"github.com/go-chi/chi/v5"

	"familyshare/internal/config"
	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/web"
)

// Helper to create a valid WebP file for testing rotation
func createTestWebP(t *testing.T, path string, w, h int) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer f.Close()

	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// Add some color to identify orientation if needed, but dimensions are enough for basic test
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})

	if err := webp.Encode(f, img, &webp.Options{Quality: 80}); err != nil {
		t.Fatalf("failed to encode webp: %v", err)
	}
}

func TestAdminRotatePhoto_Success(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	tempDir := t.TempDir()
	store := storage.New(tempDir)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10}, nil)
	q := sqlc.New(dbConn)

	// Create album
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAlbum: %v", err)
	}

	// Create photo in DB (Original: 100x50)
	photo, err := q.CreatePhoto(context.Background(), sqlc.CreatePhotoParams{
		AlbumID:   album.ID,
		Filename:  "test_rotate.webp",
		Width:     100,
		Height:    50,
		SizeBytes: 1024,
		Format:    "webp",
	})
	if err != nil {
		t.Fatalf("CreatePhoto: %v", err)
	}

	// Create real file on disk
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}

	photoPath := storage.PhotoPathAt(store.BaseDir, album.ID, photo.ID, "webp", createdAt)
	createTestWebP(t, photoPath, 100, 50)

	// Prepare Request: Rotate 90 degrees
	req := httptest.NewRequest("POST", "/admin/photos/"+strconv.FormatInt(photo.ID, 10)+"/rotate", strings.NewReader("angle=90"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Setup Route Context (chi)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(photo.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()
	h.AdminRotatePhoto(w, req)

	// Verify Status
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	if w.Header().Get("HX-Refresh") != "true" {
		t.Errorf("expected HX-Refresh header to be true")
	}

	// Verify DB Update
	updatedPhoto, err := q.GetPhoto(context.Background(), photo.ID)
	if err != nil {
		t.Fatalf("GetPhoto: %v", err)
	}

	if updatedPhoto.Width != 50 || updatedPhoto.Height != 100 {
		t.Errorf("expected dimensions 50x100, got %dx%d", updatedPhoto.Width, updatedPhoto.Height)
	}

	// Verify File on Disk
	f, err := os.Open(photoPath)
	if err != nil {
		t.Fatalf("failed to open rotated file: %v", err)
	}
	defer f.Close()
	cfg, err := webp.DecodeConfig(f)
	if err != nil {
		t.Fatalf("failed to decode config: %v", err)
	}
	if cfg.Width != 50 || cfg.Height != 100 {
		t.Errorf("file dimensions = %dx%d, want 50x100", cfg.Width, cfg.Height)
	}
}

func TestAdminRotatePhoto_InvalidAngle(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10}, nil)

	req := httptest.NewRequest("POST", "/admin/photos/1/rotate", strings.NewReader("angle=45"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()
	h.AdminRotatePhoto(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for invalid angle, got %d", w.Code)
	}
}
