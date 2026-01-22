package handler_test

import (
"familyshare/internal/config"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/web"
)

func TestCreateShareLink_WithToken(t *testing.T) {
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

	// Create share link
	vals := url.Values{}
	vals.Set("target_type", "album")
	vals.Set("target_id", "1")
	vals.Set("max_views", "10")
	vals.Set("expires_at", "2026-12-31T23:59")

	req := httptest.NewRequest("POST", "/admin/shares", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CreateShareLink(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify share link in database
	shares, err := q.ListShareLinks(context.Background(), sqlc.ListShareLinksParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListShareLinks: %v", err)
	}

	if len(shares) != 1 {
		t.Fatalf("expected 1 share link, got %d", len(shares))
	}

	share := shares[0]
	if share.TargetType != "album" {
		t.Fatalf("expected target_type 'album', got %s", share.TargetType)
	}
	if share.TargetID != album.ID {
		t.Fatalf("expected target_id %d, got %d", album.ID, share.TargetID)
	}
	if !share.MaxViews.Valid || share.MaxViews.Int64 != 10 {
		t.Fatalf("expected max_views 10, got %v", share.MaxViews)
	}
	if !share.ExpiresAt.Valid {
		t.Fatalf("expected expires_at to be set")
	}
	if share.Token == "" {
		t.Fatalf("expected token to be generated")
	}
	if len(share.Token) < 40 {
		t.Fatalf("token seems too short: %s", share.Token)
	}
}

func TestCreateShareLink_UnlimitedViews(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album
	_, err = q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateAlbum: %v", err)
	}

	// Create share link without max_views
	vals := url.Values{}
	vals.Set("target_type", "album")
	vals.Set("target_id", "1")

	req := httptest.NewRequest("POST", "/admin/shares", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CreateShareLink(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Verify max_views is null
	shares, _ := q.ListShareLinks(context.Background(), sqlc.ListShareLinksParams{Limit: 10, Offset: 0})
	if len(shares) != 1 {
		t.Fatalf("expected 1 share link, got %d", len(shares))
	}
	if shares[0].MaxViews.Valid {
		t.Fatalf("expected max_views to be null, got %v", shares[0].MaxViews)
	}
}

func TestCreateShareLink_InvalidTargetType(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	vals := url.Values{}
	vals.Set("target_type", "invalid")
	vals.Set("target_id", "1")

	req := httptest.NewRequest("POST", "/admin/shares", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CreateShareLink(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestCreateShareLink_TargetNotFound(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	vals := url.Values{}
	vals.Set("target_type", "album")
	vals.Set("target_id", "9999") // Non-existent

	req := httptest.NewRequest("POST", "/admin/shares", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CreateShareLink(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestListShareLinks_ShowsCreatedLinks(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album
	album, _ := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})

	// Create share link directly in DB
	_, err = q.CreateShareLink(context.Background(), sqlc.CreateShareLinkParams{
		Token:      "test-token-123",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{},
		ExpiresAt:  sql.NullTime{},
	})
	if err != nil {
		t.Fatalf("CreateShareLink: %v", err)
	}

	// List share links via handler
	req := httptest.NewRequest("GET", "/admin/shares", nil)
	w := httptest.NewRecorder()
	h.ListShareLinks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Check response contains the token
	if !strings.Contains(w.Body.String(), "test-token-123") {
		t.Fatalf("expected response to contain token")
	}
}

func TestRevokeShareLink_SetsRevokedAt(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album
	album, _ := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})

	// Create share link
	share, err := q.CreateShareLink(context.Background(), sqlc.CreateShareLinkParams{
		Token:      "test-token-123",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{},
		ExpiresAt:  sql.NullTime{},
	})
	if err != nil {
		t.Fatalf("CreateShareLink: %v", err)
	}

	// Revoke via handler
	req := httptest.NewRequest("DELETE", "/admin/shares/1", nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.RevokeShareLink(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}

	// Verify revoked_at is set
	updated, err := q.GetShareLink(context.Background(), share.ID)
	if err != nil {
		t.Fatalf("GetShareLink: %v", err)
	}
	if !updated.RevokedAt.Valid {
		t.Fatalf("expected revoked_at to be set")
	}
}

func TestRevokeShareLink_NotFound(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	req := httptest.NewRequest("DELETE", "/admin/shares/9999", nil)
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", "9999")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))
	w := httptest.NewRecorder()
	h.RevokeShareLink(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestCreateShareLink_ForPhoto(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New(t.TempDir())
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})
	q := sqlc.New(dbConn)

	// Create album and photo
	album, _ := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})
	photo, _ := q.CreatePhoto(context.Background(), sqlc.CreatePhotoParams{
		AlbumID:   album.ID,
		Filename:  "test.webp",
		Width:     1920,
		Height:    1080,
		SizeBytes: 50000,
		Format:    "webp",
	})

	// Create share link for photo
	vals := url.Values{}
	vals.Set("target_type", "photo")
	vals.Set("target_id", "1")

	req := httptest.NewRequest("POST", "/admin/shares", strings.NewReader(vals.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	h.CreateShareLink(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Verify share link
	shares, _ := q.ListShareLinks(context.Background(), sqlc.ListShareLinksParams{Limit: 10, Offset: 0})
	if len(shares) != 1 {
		t.Fatalf("expected 1 share link, got %d", len(shares))
	}
	if shares[0].TargetType != "photo" {
		t.Fatalf("expected target_type 'photo', got %s", shares[0].TargetType)
	}
	if shares[0].TargetID != photo.ID {
		t.Fatalf("expected target_id %d, got %d", photo.ID, shares[0].TargetID)
	}
}

func TestListActiveShareLinks_ExcludesExpired(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	q := sqlc.New(dbConn)

	// Create album
	album, _ := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{String: "Test", Valid: true},
	})

	// Create expired share link
	expiredTime := time.Now().Add(-24 * time.Hour)
	_, err = q.CreateShareLink(context.Background(), sqlc.CreateShareLinkParams{
		Token:      "expired-token",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{},
		ExpiresAt:  sql.NullTime{Time: expiredTime, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateShareLink: %v", err)
	}

	// Create active share link
	_, err = q.CreateShareLink(context.Background(), sqlc.CreateShareLinkParams{
		Token:      "active-token",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{},
		ExpiresAt:  sql.NullTime{},
	})
	if err != nil {
		t.Fatalf("CreateShareLink: %v", err)
	}

	// List active share links
	activeShares, err := q.ListActiveShareLinks(context.Background(), sqlc.ListActiveShareLinksParams{Limit: 10, Offset: 0})
	if err != nil {
		t.Fatalf("ListActiveShareLinks: %v", err)
	}

	// Should only include active link
	if len(activeShares) != 1 {
		t.Fatalf("expected 1 active share link, got %d", len(activeShares))
	}
	if activeShares[0].Token != "active-token" {
		t.Fatalf("expected active-token, got %s", activeShares[0].Token)
	}
}
