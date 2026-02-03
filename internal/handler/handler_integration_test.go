package handler_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"familyshare/internal/config"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/security"
	"familyshare/internal/storage"
	"familyshare/internal/testutil"
	"familyshare/web"

	"github.com/go-chi/chi/v5"
)

// TestAdminLogin_Integration tests the full admin authentication flow
func TestAdminLogin_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	// Create handler with test config
	cfg := &config.Config{
		AdminPasswordHash: testutil.HashPassword(t, "testpassword123"),
		DataDir:           storageDir,
	}

	h := handler.New(db, storage.New(storageDir), web.EmbedFS, cfg)

	// Test invalid password
	t.Run("InvalidPassword", func(t *testing.T) {
		form := url.Values{}
		form.Set("password", "wrongpassword")

		req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.Login(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected redirect, got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if !strings.Contains(location, "error=") {
			t.Error("expected error in redirect URL")
		}
	})

	// Test valid password creates session
	t.Run("ValidPassword", func(t *testing.T) {
		form := url.Values{}
		form.Set("password", "testpassword123")

		req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		h.Login(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected redirect, got %d", resp.StatusCode)
		}

		// Check session cookie was set
		cookies := resp.Cookies()
		var sessionCookie *http.Cookie
		for _, c := range cookies {
			if c.Name == "session_id" {
				sessionCookie = c
				break
			}
		}

		if sessionCookie == nil {
			t.Fatal("expected session cookie to be set")
		}

		if sessionCookie.Value == "" {
			t.Error("session cookie value is empty")
		}

		// Verify session exists in database
		session, err := q.GetSession(context.Background(), sessionCookie.Value)
		if err != nil {
			t.Fatalf("session not found in database: %v", err)
		}

		if session.ID != sessionCookie.Value {
			t.Errorf("session token mismatch: expected %s, got %s", sessionCookie.Value, session.ID)
		}

		// Verify expiration is in the future
		if session.ExpiresAt.Before(time.Now().UTC()) {
			t.Error("session already expired")
		}
	})
}

// TestShareLink_Integration tests share link creation and access flow
func TestShareLink_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	cfg := &config.Config{
		AdminPasswordHash: testutil.HashPassword(t, "test123"),
		DataDir:           storageDir,
	}

	h := handler.New(db, storage.New(storageDir), web.EmbedFS, cfg)

	ctx := context.Background()

	// Create test album
	album := testutil.CreateTestAlbum(t, q, "Shared Album", "Test album for sharing")
	_ = testutil.CreateTestPhoto(t, q, album.ID, "test.webp")

	// Create share link with view limit
	token, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	shareLink := testutil.CreateTestShareLink(t, q, album.ID, token, 5, expiresAt)

	t.Run("AccessValidShareLink", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/s/"+token, nil)
		w := httptest.NewRecorder()

		// Mock chi URLParam
		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("token", token)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

		h.ViewShareLink(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			t.Errorf("expected 200 OK, got %d. Body: %s", resp.StatusCode, string(body))
		}

		// Verify view was tracked
		uniqueViews, err := q.CountUniqueShareLinkViews(ctx, shareLink.ID)
		if err != nil {
			t.Fatalf("failed to count views: %v", err)
		}

		if uniqueViews < 1 {
			t.Errorf("expected at least 1 unique view, got %d", uniqueViews)
		}
	})

	t.Run("ViewLimitExceeded", func(t *testing.T) {
		// Add views up to max limit
		for i := int64(0); i < shareLink.MaxViews.Int64; i++ {
			viewerHash := fmt.Sprintf("viewer-%d", i)
			err := q.IncrementShareLinkView(ctx, sqlc.IncrementShareLinkViewParams{
				ShareLinkID: shareLink.ID,
				ViewerHash:  viewerHash,
			})
			if err != nil {
				t.Fatalf("failed to increment view: %v", err)
			}
		}

		req := httptest.NewRequest("GET", "/s/"+token, nil)
		w := httptest.NewRecorder()

		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("token", token)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

		h.ViewShareLink(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusGone {
			t.Errorf("expected 410 Gone for exceeded views, got %d", resp.StatusCode)
		}
	})

	t.Run("ExpiredLink", func(t *testing.T) {
		// Create expired link
		expiredToken, _ := security.GenerateSecureToken()
		pastTime := time.Now().UTC().Add(-24 * time.Hour)
		testutil.CreateTestShareLink(t, q, album.ID, expiredToken, 0, pastTime)

		req := httptest.NewRequest("GET", "/s/"+expiredToken, nil)
		w := httptest.NewRecorder()

		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("token", expiredToken)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

		h.ViewShareLink(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusGone {
			t.Errorf("expected 410 Gone for expired link, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/s/invalidtoken123", nil)
		w := httptest.NewRecorder()

		chiCtx := chi.NewRouteContext()
		chiCtx.URLParams.Add("token", "invalidtoken123")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

		h.ViewShareLink(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404 Not Found, got %d", resp.StatusCode)
		}
	})
}

// TestAlbumCRUD_Integration tests album creation, update, and deletion
func TestAlbumCRUD_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	_, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	ctx := context.Background()

	t.Run("CreateAlbum", func(t *testing.T) {
		album := testutil.CreateTestAlbum(t, q, "New Album", "Test Description")

		if album.Title != "New Album" {
			t.Errorf("expected title='New Album', got '%s'", album.Title)
		}

		if !album.Description.Valid || album.Description.String != "Test Description" {
			t.Error("description not saved correctly")
		}

		// Verify timestamps are in UTC
		if album.CreatedAt.Valid && album.CreatedAt.Time.Location() != time.UTC {
			t.Errorf("expected created_at in UTC, got %v", album.CreatedAt.Time.Location())
		}

		if album.UpdatedAt.Valid && album.UpdatedAt.Time.Location() != time.UTC {
			t.Errorf("expected updated_at in UTC, got %v", album.UpdatedAt.Time.Location())
		}
	})

	t.Run("UpdateAlbum", func(t *testing.T) {
		album := testutil.CreateTestAlbum(t, q, "Original Title", "Original Desc")

		// Sleep to ensure CURRENT_TIMESTAMP will be different
		time.Sleep(1100 * time.Millisecond)

		// Update the album
		err := q.UpdateAlbum(ctx, sqlc.UpdateAlbumParams{
			ID:           album.ID,
			Title:        "Updated Title",
			Description:  sql.NullString{String: "Updated Desc", Valid: true},
			CoverPhotoID: sql.NullInt64{},
		})
		if err != nil {
			t.Fatalf("failed to update album: %v", err)
		}

		// Retrieve and verify
		updated, err := q.GetAlbum(ctx, album.ID)
		if err != nil {
			t.Fatalf("failed to get updated album: %v", err)
		}

		if updated.Title != "Updated Title" {
			t.Errorf("title not updated: got '%s'", updated.Title)
		}

		if updated.UpdatedAt.Valid && album.UpdatedAt.Valid {
			if updated.UpdatedAt.Time.Before(album.UpdatedAt.Time) || updated.UpdatedAt.Time.Equal(album.UpdatedAt.Time) {
				t.Error("updated_at timestamp not changed")
			}
		}
	})

	t.Run("DeleteAlbum_CascadesToPhotos", func(t *testing.T) {
		album := testutil.CreateTestAlbum(t, q, "Delete Test", "")
		photo1 := testutil.CreateTestPhoto(t, q, album.ID, "photo1.webp")
		photo2 := testutil.CreateTestPhoto(t, q, album.ID, "photo2.webp")

		// Verify photos exist
		photos, err := q.ListPhotosByAlbum(ctx, sqlc.ListPhotosByAlbumParams{
			AlbumID: album.ID,
			Limit:   100,
			Offset:  0,
		})
		if err != nil {
			t.Fatalf("failed to list photos: %v", err)
		}
		if len(photos) != 2 {
			t.Fatalf("expected 2 photos, got %d", len(photos))
		}

		// Delete album
		err = q.DeleteAlbum(ctx, album.ID)
		if err != nil {
			t.Fatalf("failed to delete album: %v", err)
		}

		// Verify photos were cascaded
		_, err = q.GetPhoto(ctx, photo1.ID)
		if err != sql.ErrNoRows {
			t.Error("expected photo1 to be deleted via cascade")
		}

		_, err = q.GetPhoto(ctx, photo2.ID)
		if err != sql.ErrNoRows {
			t.Error("expected photo2 to be deleted via cascade")
		}
	})
}

// TestUploadAndPhotoAccess_Integration tests photo upload and retrieval
func TestUploadAndPhotoAccess_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	_, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	os.Setenv("STORAGE_PATH", storageDir)
	defer os.Unsetenv("STORAGE_PATH")

	ctx := context.Background()

	album := testutil.CreateTestAlbum(t, q, "Upload Test", "")

	t.Run("PhotoMetadata", func(t *testing.T) {
		photo := testutil.CreateTestPhoto(t, q, album.ID, "metadata_test.webp")

		// Verify all metadata fields
		if photo.AlbumID != album.ID {
			t.Errorf("album_id mismatch: expected %d, got %d", album.ID, photo.AlbumID)
		}

		if photo.Width != 1920 || photo.Height != 1080 {
			t.Errorf("dimensions mismatch: expected 1920x1080, got %dx%d", photo.Width, photo.Height)
		}

		if photo.Format != "webp" {
			t.Errorf("format mismatch: expected webp, got %s", photo.Format)
		}

		if photo.SizeBytes <= 0 {
			t.Error("size_bytes should be positive")
		}
	})

	t.Run("ListPhotosByAlbum", func(t *testing.T) {
		// Create multiple photos
		testutil.CreateTestPhoto(t, q, album.ID, "photo1.webp")
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		testutil.CreateTestPhoto(t, q, album.ID, "photo2.webp")
		time.Sleep(10 * time.Millisecond)
		testutil.CreateTestPhoto(t, q, album.ID, "photo3.webp")

		photos, err := q.ListPhotosByAlbum(ctx, sqlc.ListPhotosByAlbumParams{
			AlbumID: album.ID,
			Limit:   100,
			Offset:  0,
		})
		if err != nil {
			t.Fatalf("failed to list photos: %v", err)
		}

		if len(photos) < 3 {
			t.Errorf("expected at least 3 photos, got %d", len(photos))
		}

		// Verify photos are ordered by created_at DESC (newest first)
		for i := 0; i < len(photos)-1; i++ {
			if photos[i].CreatedAt.Valid && photos[i+1].CreatedAt.Valid {
				if photos[i].CreatedAt.Time.Before(photos[i+1].CreatedAt.Time) {
					t.Error("photos not ordered by created_at DESC")
					break
				}
			}
		}
	})
}
