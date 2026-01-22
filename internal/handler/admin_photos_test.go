package handler_test

import (
"familyshare/internal/config"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/web"
)

func TestDeletePhoto_RemovesFileAndDBRecord(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	// Create temp directory for test photos
	tempDir := t.TempDir()
	store := storage.New(tempDir)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAlbum: %v", err)
	}

	// Create photo in DB
	photo, err := q.CreatePhoto(context.Background(), sqlc.CreatePhotoParams{
		AlbumID:   album.ID,
		Filename:  "test.webp",
		Width:     1920,
		Height:    1080,
		SizeBytes: 50000,
		Format:    "webp",
	})
	if err != nil {
		t.Fatalf("CreatePhoto: %v", err)
	}

	// Create actual photo file on disk
	photoPath := storage.PhotoPath(tempDir, album.ID, photo.ID, "webp")
	if err := os.MkdirAll(filepath.Dir(photoPath), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(photoPath, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(photoPath); os.IsNotExist(err) {
		t.Fatalf("photo file was not created")
	}

	// Delete photo via handler
	req := httptest.NewRequest("DELETE", "/admin/photos/"+strconv.FormatInt(photo.ID, 10), nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(photo.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.DeletePhoto(w, req)

	// Verify response
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}

	// Verify photo removed from DB
	if _, err := q.GetPhoto(context.Background(), photo.ID); err == nil {
		t.Fatalf("expected photo to be deleted from DB")
	}

	// Verify file removed from disk
	if _, err := os.Stat(photoPath); !os.IsNotExist(err) {
		t.Fatalf("expected photo file to be deleted, but it still exists")
	}
}

func TestDeletePhoto_NotFound(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	// Try to delete non-existent photo
	req := httptest.NewRequest("DELETE", "/admin/photos/99999", nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.DeletePhoto(w, req)

	// Verify 404 response
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestDeletePhoto_InvalidID(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	// Try to delete with invalid ID
	req := httptest.NewRequest("DELETE", "/admin/photos/invalid", nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.DeletePhoto(w, req)

	// Verify 400 response
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDeletePhoto_FileNotExist_SucceedsDBDeletion(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	tempDir := t.TempDir()
	store := storage.New(tempDir)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAlbum: %v", err)
	}

	// Create photo in DB (but don't create file)
	photo, err := q.CreatePhoto(context.Background(), sqlc.CreatePhotoParams{
		AlbumID:   album.ID,
		Filename:  "missing.webp",
		Width:     1920,
		Height:    1080,
		SizeBytes: 50000,
		Format:    "webp",
	})
	if err != nil {
		t.Fatalf("CreatePhoto: %v", err)
	}

	// Delete photo via handler (file doesn't exist, but should succeed)
	req := httptest.NewRequest("DELETE", "/admin/photos/"+strconv.FormatInt(photo.ID, 10), nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(photo.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.DeletePhoto(w, req)

	// Should succeed even if file doesn't exist
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}

	// Verify photo removed from DB
	if _, err := q.GetPhoto(context.Background(), photo.ID); err == nil {
		t.Fatalf("expected photo to be deleted from DB")
	}
}

func TestSetCoverPhoto_UpdatesAlbum(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAlbum: %v", err)
	}

	// Create photo
	photo, err := q.CreatePhoto(context.Background(), sqlc.CreatePhotoParams{
		AlbumID:   album.ID,
		Filename:  "cover.webp",
		Width:     1920,
		Height:    1080,
		SizeBytes: 50000,
		Format:    "webp",
	})
	if err != nil {
		t.Fatalf("CreatePhoto: %v", err)
	}

	// Set as cover photo
	req := httptest.NewRequest("POST", "/admin/photos/"+strconv.FormatInt(photo.ID, 10)+"/set-cover", nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(photo.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.SetCoverPhoto(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Verify HTMX trigger header
	if w.Header().Get("HX-Trigger") != "coverPhotoUpdated" {
		t.Fatalf("expected HX-Trigger header, got %s", w.Header().Get("HX-Trigger"))
	}

	// Verify album cover updated
	updatedAlbum, err := q.GetAlbum(context.Background(), album.ID)
	if err != nil {
		t.Fatalf("GetAlbum: %v", err)
	}
	if !updatedAlbum.CoverPhotoID.Valid || updatedAlbum.CoverPhotoID.Int64 != photo.ID {
		t.Fatalf("expected cover photo ID %d, got %v", photo.ID, updatedAlbum.CoverPhotoID)
	}
}

func TestSetCoverPhoto_PhotoNotFound(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	// Try to set non-existent photo as cover
	req := httptest.NewRequest("POST", "/admin/photos/99999/set-cover", nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", "99999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.SetCoverPhoto(w, req)

	// Verify 404 response
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestSetCoverPhoto_InvalidID(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	// Try to set cover with invalid ID
	req := httptest.NewRequest("POST", "/admin/photos/invalid/set-cover", nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.SetCoverPhoto(w, req)

	// Verify 400 response
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDeletePhoto_ThatIsAlbumCover_ClearsAlbumCover(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	tempDir := t.TempDir()
	store := storage.New(tempDir)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAlbum: %v", err)
	}

	// Create photo
	photo, err := q.CreatePhoto(context.Background(), sqlc.CreatePhotoParams{
		AlbumID:   album.ID,
		Filename:  "cover.webp",
		Width:     1920,
		Height:    1080,
		SizeBytes: 50000,
		Format:    "webp",
	})
	if err != nil {
		t.Fatalf("CreatePhoto: %v", err)
	}

	// Create photo file
	photoPath := storage.PhotoPath(tempDir, album.ID, photo.ID, "webp")
	if err := os.MkdirAll(filepath.Dir(photoPath), 0755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(photoPath, []byte("fake image data"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Set as cover photo
	if err := q.SetAlbumCover(context.Background(), sqlc.SetAlbumCoverParams{
		CoverPhotoID: sql.NullInt64{Int64: photo.ID, Valid: true},
		ID:           album.ID,
	}); err != nil {
		t.Fatalf("SetAlbumCover: %v", err)
	}

	// Verify cover is set
	albumWithCover, err := q.GetAlbum(context.Background(), album.ID)
	if err != nil {
		t.Fatalf("GetAlbum: %v", err)
	}
	if !albumWithCover.CoverPhotoID.Valid {
		t.Fatalf("expected cover photo to be set")
	}

	// Delete the cover photo
	req := httptest.NewRequest("DELETE", "/admin/photos/"+strconv.FormatInt(photo.ID, 10), nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(photo.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.DeletePhoto(w, req)

	// Verify response
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}

	// Verify album cover is now null (due to ON DELETE SET NULL in schema)
	albumAfterDelete, err := q.GetAlbum(context.Background(), album.ID)
	if err != nil {
		t.Fatalf("GetAlbum: %v", err)
	}
	if albumAfterDelete.CoverPhotoID.Valid {
		t.Fatalf("expected cover photo to be cleared (null) after deletion, but got %v", albumAfterDelete.CoverPhotoID)
	}
}
