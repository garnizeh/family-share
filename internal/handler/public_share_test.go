package handler_test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/config"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/internal/testutil"
	"familyshare/web"
)

func setupTestHandlerForShare(t *testing.T) (*handler.Handler, *sqlc.Queries, func()) {
	dbConn, q, dbCleanup := testutil.SetupTestDB(t)

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	t.Setenv("STORAGE_PATH", storageDir)
	store := storage.New(storageDir)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	cleanup := func() {
		storageCleanup()
		dbCleanup()
	}

	return h, q, cleanup
}

func TestViewShareLink_NotFound(t *testing.T) {
	h, _, cleanup := setupTestHandlerForShare(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/s/nonexistent-token", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "nonexistent-token")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ViewShareLink(w, req)

	// Should return 404 with expired page
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Expected error page content")
	}
}

func TestViewShareLink_Revoked(t *testing.T) {
	h, q, cleanup := setupTestHandlerForShare(t)
	defer cleanup()

	ctx := context.Background()

	// Create test album
	album, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test album: %v", err)
	}

	// Create share link
	link, err := q.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "test-revoked-token",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{Valid: false},
		ExpiresAt:  sql.NullTime{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create share link: %v", err)
	}

	// Revoke it
	err = q.RevokeShareLink(ctx, link.ID)
	if err != nil {
		t.Fatalf("Failed to revoke share link: %v", err)
	}

	// Try to access
	req := httptest.NewRequest("GET", "/s/test-revoked-token", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "test-revoked-token")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ViewShareLink(w, req)

	// Should return 410 Gone
	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410 Gone, got %d", w.Code)
	}
}

func TestViewShareLink_Expired(t *testing.T) {
	h, q, cleanup := setupTestHandlerForShare(t)
	defer cleanup()

	ctx := context.Background()

	// Create test album
	album, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test album: %v", err)
	}

	// Create expired share link (expired 1 hour ago)
	expiredTime := time.Now().Add(-1 * time.Hour)
	_, err = q.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "test-expired-token",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{Valid: false},
		ExpiresAt:  sql.NullTime{Valid: true, Time: expiredTime},
	})
	if err != nil {
		t.Fatalf("Failed to create share link: %v", err)
	}

	// Try to access
	req := httptest.NewRequest("GET", "/s/test-expired-token", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "test-expired-token")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ViewShareLink(w, req)

	// Should return 410 Gone
	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410 Gone for expired link, got %d", w.Code)
	}
}

func TestViewShareLink_ViewLimitReached(t *testing.T) {
	h, q, cleanup := setupTestHandlerForShare(t)
	defer cleanup()

	ctx := context.Background()

	// Create test album
	album, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test album: %v", err)
	}

	// Create share link with max 2 views
	link, err := q.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "test-limited-token",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{Valid: true, Int64: 2},
		ExpiresAt:  sql.NullTime{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create share link: %v", err)
	}

	// Simulate 2 unique viewers
	err = q.IncrementShareLinkView(ctx, sqlc.IncrementShareLinkViewParams{
		ShareLinkID: link.ID,
		ViewerHash:  "hash1",
	})
	if err != nil {
		t.Fatalf("Failed to increment view 1: %v", err)
	}

	err = q.IncrementShareLinkView(ctx, sqlc.IncrementShareLinkViewParams{
		ShareLinkID: link.ID,
		ViewerHash:  "hash2",
	})
	if err != nil {
		t.Fatalf("Failed to increment view 2: %v", err)
	}

	// Try to access (3rd viewer)
	req := httptest.NewRequest("GET", "/s/test-limited-token", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "test-limited-token")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ViewShareLink(w, req)

	// Should return 410 Gone (limit reached)
	if w.Code != http.StatusGone {
		t.Errorf("Expected status 410 Gone for view limit, got %d", w.Code)
	}
}

func TestViewShareLink_ValidAlbum_FirstVisit(t *testing.T) {
	h, q, cleanup := setupTestHandlerForShare(t)
	defer cleanup()

	ctx := context.Background()

	// Create test album
	album, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{Valid: true, String: "A test album"},
	})
	if err != nil {
		t.Fatalf("Failed to create test album: %v", err)
	}

	// Create share link
	link, err := q.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "test-valid-token-abc123",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{Valid: false},
		ExpiresAt:  sql.NullTime{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create share link: %v", err)
	}

	// Access the link
	req := httptest.NewRequest("GET", "/s/test-valid-token-abc123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "test-valid-token-abc123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ViewShareLink(w, req)

	// Should return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check that view was tracked
	count, err := q.CountUniqueShareLinkViews(ctx, link.ID)
	if err != nil {
		t.Fatalf("Failed to count views: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 unique view, got %d", count)
	}

	// Check cookie was set
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "_vh_test-val" {
			found = true
			if cookie.Value == "" {
				t.Error("Expected non-empty viewer hash cookie")
			}
			if !cookie.HttpOnly {
				t.Error("Expected HttpOnly cookie")
			}
		}
	}
	if !found {
		t.Error("Expected viewer hash cookie to be set")
	}
}

func TestViewShareLink_ValidAlbum_SecondVisit(t *testing.T) {
	h, q, cleanup := setupTestHandlerForShare(t)
	defer cleanup()

	ctx := context.Background()

	// Create test album
	album, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test album: %v", err)
	}

	// Create share link
	link, err := q.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "test-repeat-token-xyz789",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{Valid: false},
		ExpiresAt:  sql.NullTime{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create share link: %v", err)
	}

	// First visit
	req1 := httptest.NewRequest("GET", "/s/test-repeat-token-xyz789", nil)
	req1.RemoteAddr = "192.168.1.100:12345"
	req1.Header.Set("User-Agent", "TestBrowser/1.0")
	rctx1 := chi.NewRouteContext()
	rctx1.URLParams.Add("token", "test-repeat-token-xyz789")
	req1 = req1.WithContext(context.WithValue(req1.Context(), chi.RouteCtxKey, rctx1))
	w1 := httptest.NewRecorder()
	h.ViewShareLink(w1, req1)

	// Get the cookie
	var viewerCookie *http.Cookie
	for _, cookie := range w1.Result().Cookies() {
		if cookie.Name == "_vh_test-rep" {
			viewerCookie = cookie
			break
		}
	}
	if viewerCookie == nil {
		t.Fatal("Expected viewer cookie to be set on first visit")
	}

	// Second visit with same cookie
	req2 := httptest.NewRequest("GET", "/s/test-repeat-token-xyz789", nil)
	req2.RemoteAddr = "192.168.1.100:12345"
	req2.Header.Set("User-Agent", "TestBrowser/1.0")
	req2.AddCookie(viewerCookie)
	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("token", "test-repeat-token-xyz789")
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rctx2))
	w2 := httptest.NewRecorder()
	h.ViewShareLink(w2, req2)

	// Both should succeed
	if w1.Code != http.StatusOK || w2.Code != http.StatusOK {
		t.Errorf("Expected both visits to return 200, got %d and %d", w1.Code, w2.Code)
	}

	// But view count should still be 1 (same viewer)
	count, err := q.CountUniqueShareLinkViews(ctx, link.ID)
	if err != nil {
		t.Fatalf("Failed to count views: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 unique view after repeat visit, got %d", count)
	}
}

func TestViewShareLink_ValidPhoto(t *testing.T) {
	h, q, cleanup := setupTestHandlerForShare(t)
	defer cleanup()

	ctx := context.Background()

	// Create test album
	album, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Test Album",
		Description: sql.NullString{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create test album: %v", err)
	}

	// Create test photo
	photo, err := q.CreatePhoto(ctx, sqlc.CreatePhotoParams{
		AlbumID:   album.ID,
		Filename:  "test.jpg",
		SizeBytes: 12345,
		Width:     800,
		Height:    600,
		Format:    "webp",
	})
	if err != nil {
		t.Fatalf("Failed to create test photo: %v", err)
	}

	// Create share link for photo
	_, err = q.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "test-photo-token-123",
		TargetType: "photo",
		TargetID:   photo.ID,
		MaxViews:   sql.NullInt64{Valid: false},
		ExpiresAt:  sql.NullTime{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create share link: %v", err)
	}

	// Access the link
	req := httptest.NewRequest("GET", "/s/test-photo-token-123", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "test-photo-token-123")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.ViewShareLink(w, req)

	// Should return 200 OK
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Body should contain photo template content
	body := w.Body.String()
	if body == "" {
		t.Error("Expected response body")
	}
}

func TestViewShareLink_AlbumPagination(t *testing.T) {
	h, q, cleanup := setupTestHandlerForShare(t)
	defer cleanup()

	ctx := context.Background()

	// Create test album
	album, err := q.CreateAlbum(ctx, sqlc.CreateAlbumParams{
		Title:       "Large Album",
		Description: sql.NullString{String: "Album with many photos", Valid: true},
	})
	if err != nil {
		t.Fatalf("Failed to create test album: %v", err)
	}

	// Create 25 photos to test pagination (20 per page)
	for i := 1; i <= 25; i++ {
		_, err := q.CreatePhoto(ctx, sqlc.CreatePhotoParams{
			AlbumID:   album.ID,
			Filename:  fmt.Sprintf("photo%d.webp", i),
			Width:     1920,
			Height:    1080,
			SizeBytes: 100000,
			Format:    "webp",
		})
		if err != nil {
			t.Fatalf("Failed to create photo %d: %v", i, err)
		}
	}

	// Create share link
	_, err = q.CreateShareLink(ctx, sqlc.CreateShareLinkParams{
		Token:      "test-pagination-token",
		TargetType: "album",
		TargetID:   album.ID,
		MaxViews:   sql.NullInt64{Valid: false},
		ExpiresAt:  sql.NullTime{Valid: false},
	})
	if err != nil {
		t.Fatalf("Failed to create share link: %v", err)
	}

	// Test first page
	t.Run("first page shows load more button", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/s/test-pagination-token", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "test-pagination-token")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.ViewShareLink(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		// Should contain Load More button
		if !contains(body, "Load More Photos") {
			t.Error("Expected to find Load More button on first page")
		}
		// Should have photo grid
		if !contains(body, "photo-grid") {
			t.Error("Expected to find photo grid")
		}
	})

	// Test HTMX partial request
	t.Run("HTMX request returns partial HTML", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/s/test-pagination-token?page=2", nil)
		req.Header.Set("HX-Request", "true")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "test-pagination-token")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.ViewShareLink(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		// Should NOT contain full HTML document
		if contains(body, "<!DOCTYPE html>") {
			t.Error("HTMX partial should not contain full HTML document")
		}
		// Should contain photo cards
		if !contains(body, "card-photo") {
			t.Error("Expected to find photo cards in partial")
		}
	})

	// Test last page doesn't show load more
	t.Run("last page hides load more button", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/s/test-pagination-token?page=2", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "test-pagination-token")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()

		h.ViewShareLink(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		// Should NOT have Load More button (only 25 photos, page 2 has 5 remaining)
		if contains(body, "Load More Photos") {
			t.Error("Expected NO Load More button on last page")
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
