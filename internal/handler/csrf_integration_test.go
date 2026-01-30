package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/config"
	"familyshare/internal/handler"
	"familyshare/internal/security"
	"familyshare/internal/storage"
	"familyshare/internal/testutil"
	"familyshare/web"
)

func TestAdminLogin_CSRFRequired(t *testing.T) {
	db, _, cleanup := testutil.SetupTestDB(t)
	defer cleanup()

	hash, err := security.HashPassword("secret")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	cfg := &config.Config{
		RateLimitShare:    60,
		RateLimitAdmin:    10,
		AdminPasswordHash: hash,
		CSRFSecret:        "test-secret",
	}

	h := handler.New(db, storage.New(t.TempDir()), web.EmbedFS, cfg)
	r := chi.NewRouter()
	h.RegisterRoutes(r)

	// POST without CSRF token should be forbidden
	postReq := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	postRec := httptest.NewRecorder()
	r.ServeHTTP(postRec, postReq)
	if postRec.Result().StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", postRec.Result().StatusCode)
	}

	// GET to receive CSRF cookie
	getReq := httptest.NewRequest(http.MethodGet, "/admin/login", nil)
	getRec := httptest.NewRecorder()
	r.ServeHTTP(getRec, getReq)

	cookies := getRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected csrf cookie")
	}
	csrfCookie := cookies[0]

	// POST with CSRF header should not be forbidden
	form := url.Values{}
	form.Set("password", "secret")
	postReq = httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	postReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	postReq.AddCookie(csrfCookie)
	postReq.Header.Set("X-CSRF-Token", csrfCookie.Value)
	postRec = httptest.NewRecorder()

	r.ServeHTTP(postRec, postReq)
	if postRec.Result().StatusCode == http.StatusForbidden {
		t.Fatalf("expected non-403, got %d", postRec.Result().StatusCode)
	}
}
