package handler_test

import (
"familyshare/internal/config"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/web"
)

func TestCreateAndListAndUpdateAndDeleteAlbum(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New("./testdata")
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album via handler
	vals := url.Values{}
	vals.Set("title", "My Album")
	vals.Set("description", "Desc")
	req := httptest.NewRequest("POST", "/admin/albums", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CreateAlbum(w, req)

	// Ensure album exists in DB
	albums, err := q.ListAlbums(context.Background(), sqlc.ListAlbumsParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListAlbums: %v", err)
	}
	if len(albums) != 1 {
		t.Fatalf("expected 1 album, got %d", len(albums))
	}
	if albums[0].Title != "My Album" {
		t.Fatalf("unexpected title: %s", albums[0].Title)
	}

	// Update album via handler
	id := albums[0].ID
	vals2 := url.Values{}
	vals2.Set("title", "New Title")
	vals2.Set("description", "New Desc")
	req2 := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(id, 10), strings.NewReader(vals2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(id, 10))
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rc))
	w2 := httptest.NewRecorder()
	h.UpdateAlbum(w2, req2)

	// Fetch album and verify update
	updated, err := q.GetAlbum(context.Background(), id)
	if err != nil {
		t.Fatalf("GetAlbum: %v", err)
	}
	if updated.Title != "New Title" {
		t.Fatalf("expected updated title, got %s", updated.Title)
	}

	// Add a photo to the album
	_, err = q.CreatePhoto(context.Background(), sqlc.CreatePhotoParams{AlbumID: id, Filename: "p.webp", Width: 100, Height: 100, SizeBytes: 1234, Format: "webp"})
	if err != nil {
		t.Fatalf("CreatePhoto: %v", err)
	}

	// Delete album via handler
	req3 := httptest.NewRequest("DELETE", "/admin/albums/"+strconv.FormatInt(id, 10), nil)
	rc3 := chi.NewRouteContext()
	rc3.URLParams.Add("id", strconv.FormatInt(id, 10))
	req3 = req3.WithContext(context.WithValue(req3.Context(), chi.RouteCtxKey, rc3))
	w3 := httptest.NewRecorder()
	h.DeleteAlbum(w3, req3)

	// Verify album deleted
	if _, err := q.GetAlbum(context.Background(), id); err == nil {
		t.Fatalf("expected album to be deleted")
	}

	// Verify photos cascaded
	photos, err := q.ListPhotosByAlbum(context.Background(), sqlc.ListPhotosByAlbumParams{AlbumID: id, Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListPhotosByAlbum: %v", err)
	}
	if len(photos) != 0 {
		t.Fatalf("expected photos to be deleted with album, found %d", len(photos))
	}
}

func TestCreateAlbumValidation(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New("./testdata")
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	// Missing title -> bad request
	vals := url.Values{}
	vals.Set("title", "")
	req := httptest.NewRequest("POST", "/admin/albums", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CreateAlbum(w, req)
	if w.Result().StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for empty title, got %d", w.Result().StatusCode)
	}
}
