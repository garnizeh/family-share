package handler_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"familyshare/internal/config"
	"familyshare/internal/handler"
	"familyshare/internal/security"
	"familyshare/internal/storage"
	"familyshare/internal/testutil"
	"familyshare/web"

	"github.com/go-chi/chi/v5"
)

// Test that direct /data/photos/* route is not publicly available via router
func TestDirectDataPhotosRoute_NotFound(t *testing.T) {
	db, _, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	cfg := &config.Config{DataDir: storageDir, RateLimitShare: 100000}
	h := handler.New(db, storage.New(storageDir), web.EmbedFS, cfg, nil)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/data/photos/1.webp", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 404 or 405 for direct data photo route, got %d", resp.StatusCode)
	}
}

// Test that shared photo can be fetched via /s/{token}/photos/{id}.webp
func TestServeSharedPhoto_AllowsWithValidToken(t *testing.T) {
	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	cfg := &config.Config{DataDir: storageDir, RateLimitShare: 100000}
	h := handler.New(db, storage.New(storageDir), web.EmbedFS, cfg, nil)

	// create album and photo record
	album := testutil.CreateTestAlbum(t, q, "Public Album", "")
	photo := testutil.CreateTestPhoto(t, q, album.ID, "shared.webp")

	// create a physical file at expected storage path
	ext := photo.Format
	if ext == "" {
		ext = "webp"
	}
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	path := storage.PhotoPathAt(storageDir, album.ID, photo.ID, ext, createdAt)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create photo dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("testdata"), 0644); err != nil {
		t.Fatalf("failed to write photo file: %v", err)
	}

	token, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	timeOut := time.Now().UTC().Add(1 * time.Hour)
	testutil.CreateTestShareLink(t, q, album.ID, token, 0, timeOut)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	url := "/s/" + token + "/photos/" + int64ToStr(photo.ID) + ".webp"
	req := httptest.NewRequest("GET", url, nil)
	// Ensure rate limiter sees a unique client IP for this test
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for shared photo, got %d", resp.StatusCode)
	}
}

// Test that invalid token cannot fetch photo
func TestServeSharedPhoto_DeniesWithInvalidToken(t *testing.T) {
	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	cfg := &config.Config{DataDir: storageDir, RateLimitShare: 100000}
	h := handler.New(db, storage.New(storageDir), web.EmbedFS, cfg, nil)

	album := testutil.CreateTestAlbum(t, q, "Private Album", "")
	photo := testutil.CreateTestPhoto(t, q, album.ID, "private.webp")

	// create physical file
	ext := photo.Format
	if ext == "" {
		ext = "webp"
	}
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	path := storage.PhotoPathAt(storageDir, album.ID, photo.ID, ext, createdAt)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create photo dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("testdata"), 0644); err != nil {
		t.Fatalf("failed to write photo file: %v", err)
	}

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/s/invalid-token/photos/"+int64ToStr(photo.ID)+".webp", nil)
	// Ensure rate limiter sees a unique client IP for this test
	req.Header.Set("X-Forwarded-For", "203.0.113.6")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found for invalid token, got %d", resp.StatusCode)
	}
}

// Test that shared photo path derivation uses the persisted created_at timestamp
// instead of current time (stable across month/year boundaries).
func TestServeSharedPhoto_UsesCreatedAtPath(t *testing.T) {
	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	cfg := &config.Config{DataDir: storageDir, RateLimitShare: 100000}
	h := handler.New(db, storage.New(storageDir), web.EmbedFS, cfg, nil)

	album := testutil.CreateTestAlbum(t, q, "Stable Path Album", "")
	photo := testutil.CreateTestPhoto(t, q, album.ID, "stable.webp")

	createdAt := time.Date(2025, time.December, 15, 10, 30, 0, 0, time.UTC)
	if _, err := db.Exec("UPDATE photos SET created_at = ? WHERE id = ?", createdAt, photo.ID); err != nil {
		t.Fatalf("failed to update created_at: %v", err)
	}

	path := storage.PhotoPathAt(storageDir, album.ID, photo.ID, photo.Format, createdAt)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create photo dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("testdata"), 0644); err != nil {
		t.Fatalf("failed to write photo file: %v", err)
	}

	// create share link for album
	token, err := security.GenerateSecureToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	timeOut := time.Now().UTC().Add(1 * time.Hour)
	testutil.CreateTestShareLink(t, q, album.ID, token, 0, timeOut)

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	url := "/s/" + token + "/photos/" + int64ToStr(photo.ID) + ".webp"
	req := httptest.NewRequest("GET", url, nil)
	// Ensure rate limiter sees a unique client IP for this test
	req.Header.Set("X-Forwarded-For", "203.0.113.7")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for shared photo with stable path, got %d", resp.StatusCode)
	}
}

// Test that admin photo route requires authentication
func TestAdminPhotoRoute_RequiresAuth(t *testing.T) {
	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	cfg := &config.Config{DataDir: storageDir, AdminPasswordHash: testutil.HashPassword(t, "pwd"), RateLimitAdmin: 100000, RateLimitShare: 100000}
	h := handler.New(db, storage.New(storageDir), web.EmbedFS, cfg, nil)

	album := testutil.CreateTestAlbum(t, q, "Admin Album", "")
	photo := testutil.CreateTestPhoto(t, q, album.ID, "admin.webp")

	// ensure file exists
	ext := photo.Format
	if ext == "" {
		ext = "webp"
	}
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	path := storage.PhotoPathAt(storageDir, album.ID, photo.ID, ext, createdAt)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("failed to create photo dir: %v", err)
	}
	if err := os.WriteFile(path, []byte("testdata"), 0644); err != nil {
		t.Fatalf("failed to write photo file: %v", err)
	}

	r := chi.NewRouter()
	h.RegisterRoutes(r)

	req := httptest.NewRequest("GET", "/admin/photos/"+int64ToStr(photo.ID)+".webp", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	resp := w.Result()
	// RequireAuth redirects to /admin/login (SeeOther)
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected redirect to login (303), got %d", resp.StatusCode)
	}
}

// helpers
func int64ToStr(id int64) string {
	return strconv.FormatInt(id, 10)
}
