package handler_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"familyshare/internal/config"
	"familyshare/internal/db"
	"familyshare/internal/handler"
	"familyshare/internal/security"
	"familyshare/internal/storage"
	"familyshare/web"
)

func TestLoginPage(t *testing.T) {
	h := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/admin/login", nil)
	w := httptest.NewRecorder()

	h.LoginPage(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "Admin Login") {
		t.Error("expected login page to contain 'Admin Login'")
	}
}

func TestLogin_Success(t *testing.T) {
	h := setupTestHandler(t)

	// Create a password hash for testing
	testPassword := "testpassword123"
	passwordHash, err := security.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	// Need to get config and update it via unexported field access workaround
	// For testing, we'll use a test-specific approach
	cfg := &config.Config{
		ServerAddr:        ":8080",
		DatabasePath:      ":memory:",
		DataDir:           "./testdata",
		RateLimitShare:    60,
		RateLimitAdmin:    10,
		AdminPasswordHash: passwordHash,
	}
	// Recreate handler with password hash
	dbConn, _ := db.InitDB(":memory:")
	t.Cleanup(func() { dbConn.Close() })
	store := storage.New("./testdata")
	h = handler.New(dbConn, store, web.EmbedFS, cfg)

	// Create login request
	form := url.Values{}
	form.Set("password", testPassword)
	req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Login(w, req)

	// Should redirect to admin dashboard
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/admin" {
		t.Errorf("expected redirect to /admin, got %s", location)
	}

	// Check that session cookie was set
	cookies := w.Result().Cookies()
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

	if !sessionCookie.HttpOnly {
		t.Error("session cookie should be HttpOnly")
	}

	if sessionCookie.SameSite != http.SameSiteLaxMode {
		t.Error("session cookie should have SameSite=Lax")
	}

	// We can't directly access the database from the handler in tests,
	// but the fact that we got a session cookie and successful redirect
	// indicates the session was created properly
}

func TestLogin_WrongPassword(t *testing.T) {
	// Create a password hash for testing
	correctPassword := "correctpassword"
	passwordHash, err := security.HashPassword(correctPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	cfg := &config.Config{
		ServerAddr:        ":8080",
		DatabasePath:      ":memory:",
		DataDir:           "./testdata",
		RateLimitShare:    60,
		RateLimitAdmin:    10,
		AdminPasswordHash: passwordHash,
	}
	dbConn, _ := db.InitDB(":memory:")
	t.Cleanup(func() { dbConn.Close() })
	store := storage.New("./testdata")
	h := handler.New(dbConn, store, web.EmbedFS, cfg)

	// Try to login with wrong password
	form := url.Values{}
	form.Set("password", "wrongpassword")
	req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Login(w, req)

	// Should redirect to login with error parameter
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", w.Code)
	}

	// Should redirect with error parameter
	location := w.Header().Get("Location")
	if !strings.Contains(location, "/admin/login") || !strings.Contains(location, "error=invalid_password") {
		t.Errorf("expected redirect to /admin/login?error=invalid_password, got %s", location)
	}

	// Should not set session cookie
	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "session_id" {
			t.Error("session cookie should not be set on failed login")
		}
	}
}

func TestLogin_NoPasswordConfigured(t *testing.T) {
	// Don't set AdminPasswordHash
	cfg := &config.Config{
		ServerAddr:        ":8080",
		DatabasePath:      ":memory:",
		DataDir:           "./testdata",
		RateLimitShare:    60,
		RateLimitAdmin:    10,
		AdminPasswordHash: "",
	}
	dbConn, _ := db.InitDB(":memory:")
	t.Cleanup(func() { dbConn.Close() })
	store := storage.New("./testdata")
	h := handler.New(dbConn, store, web.EmbedFS, cfg)

	form := url.Values{}
	form.Set("password", "anypassword")
	req := httptest.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	h.Login(w, req)

	// Should redirect to login with error (password verification will fail if no hash configured)
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", w.Code)
	}

	// Should redirect with error parameter
	location := w.Header().Get("Location")
	if !strings.Contains(location, "/admin/login") || !strings.Contains(location, "error=invalid_password") {
		t.Errorf("expected redirect to /admin/login with error, got %s", location)
	}
}

func TestLogout(t *testing.T) {
	// Setup handler and create a session by logging in first
	testPassword := "testpassword123"
	passwordHash, err := security.HashPassword(testPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	cfg := &config.Config{
		ServerAddr:        ":8080",
		DatabasePath:      ":memory:",
		DataDir:           "./testdata",
		RateLimitShare:    60,
		RateLimitAdmin:    10,
		AdminPasswordHash: passwordHash,
	}
	dbConn, _ := db.InitDB(":memory:")
	t.Cleanup(func() { dbConn.Close() })
	store := storage.New("./testdata")
	h := handler.New(dbConn, store, web.EmbedFS, cfg)

	// Login to create a session
	form := url.Values{}
	form.Set("password", testPassword)
	loginReq := httptest.NewRequest("POST", "/admin/login", strings.NewReader(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginW := httptest.NewRecorder()
	h.Login(loginW, loginReq)

	// Get the session cookie from login response
	var sessionCookie *http.Cookie
	for _, c := range loginW.Result().Cookies() {
		if c.Name == "session_id" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("failed to get session cookie from login")
	}

	// Create logout request with session cookie
	req := httptest.NewRequest("POST", "/admin/logout", nil)
	req.AddCookie(sessionCookie)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	// Should redirect to login page
	if w.Code != http.StatusSeeOther {
		t.Errorf("expected status 303, got %d", w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/admin/login" {
		t.Errorf("expected redirect to /admin/login, got %s", location)
	}

	// Check that session cookie was deleted
	cookies := w.Result().Cookies()
	var deletedCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "session_id" {
			deletedCookie = c
			break
		}
	}

	if deletedCookie == nil {
		t.Fatal("expected session cookie to be cleared")
	}

	if deletedCookie.MaxAge != -1 {
		t.Error("session cookie should have MaxAge=-1 to delete it")
	}
}

func TestIsValidSession(t *testing.T) {
	// This test requires access to internal handler methods
	// We'll test session validation indirectly through the protected endpoints
	// by trying to access them with and without valid sessions
	t.Skip("isValidSession is internal - tested indirectly through integration tests")
}

func setupTestHandler(t *testing.T) *handler.Handler {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	t.Cleanup(func() { dbConn.Close() })

	store := storage.New("./testdata")
	cfg := &config.Config{
		ServerAddr:     ":8080",
		DatabasePath:   ":memory:",
		DataDir:        "./testdata",
		RateLimitShare: 60,
		RateLimitAdmin: 10,
	}
	h := handler.New(dbConn, store, web.EmbedFS, cfg)
	return h
}
