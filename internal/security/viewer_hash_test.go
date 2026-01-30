package security

import (
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	_ = SetViewerHashSecret("test-secret", true)
	os.Exit(m.Run())
}

func TestGenerateViewerHash(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "test-token-123"
	ip := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	hash1 := GenerateViewerHash(token, ip, userAgent)
	hash2 := GenerateViewerHash(token, ip, userAgent)

	// Should be deterministic
	if hash1 != hash2 {
		t.Errorf("Expected same hash for same inputs, got %s and %s", hash1, hash2)
	}

	// Should be 64 characters (sha256 hex)
	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}
}

func TestGenerateViewerHash_Uniqueness(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "test-token-123"
	ip := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	hash1 := GenerateViewerHash(token, ip, userAgent)
	hash2 := GenerateViewerHash("different-token", ip, userAgent)
	hash3 := GenerateViewerHash(token, "192.168.1.2", userAgent)
	hash4 := GenerateViewerHash(token, ip, "Different User Agent")

	// All should be different
	if hash1 == hash2 || hash1 == hash3 || hash1 == hash4 {
		t.Error("Expected different hashes for different inputs")
	}
}

func TestGetViewerHash_NewVisitor(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "test-token-abc123xyz"
	req := httptest.NewRequest("GET", "/s/"+token, nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.100:12345"

	hash := GetViewerHash(req, token)

	// Should return a valid hash
	if len(hash) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash))
	}
}

func TestGetViewerHash_ExistingCookie(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "test-token-abc123xyz"
	existingHash := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	req := httptest.NewRequest("GET", "/s/"+token, nil)
	req.AddCookie(&http.Cookie{
		Name:  "_vh_test-tok",
		Value: existingHash,
	})

	hash := GetViewerHash(req, token)

	// Should return the existing hash from cookie
	if hash != existingHash {
		t.Errorf("Expected hash from cookie %s, got %s", existingHash, hash)
	}
}

func TestSetViewerHashCookie(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "test-token-abc123xyz"
	viewerHash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	w := httptest.NewRecorder()
	SetViewerHashCookie(w, token, viewerHash, nil, CookieOptions{Secure: false, SameSite: http.SameSiteLaxMode})

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != "_vh_test-tok" {
		t.Errorf("Expected cookie name '_vh_test-tok', got '%s'", cookie.Name)
	}
	if cookie.Value != viewerHash {
		t.Errorf("Expected cookie value %s, got %s", viewerHash, cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Error("Expected HttpOnly cookie")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("Expected SameSite=Lax, got %v", cookie.SameSite)
	}
	if cookie.Path != "/s/"+token {
		t.Errorf("Expected path /s/%s, got %s", token, cookie.Path)
	}
}

func TestSetViewerHashCookie_WithExpiration(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "test-token-abc123xyz"
	viewerHash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	expiresAt := time.Now().Add(2 * time.Hour)

	w := httptest.NewRecorder()
	SetViewerHashCookie(w, token, viewerHash, &expiresAt, CookieOptions{Secure: false, SameSite: http.SameSiteLaxMode})

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	// MaxAge should be approximately 2 hours (7200 seconds)
	if cookie.MaxAge < 7100 || cookie.MaxAge > 7300 {
		t.Errorf("Expected MaxAge around 7200, got %d", cookie.MaxAge)
	}
}

func TestGetClientIP_RemoteAddr(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	SetTrustedProxyCIDRs(nil)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	ip := getClientIP(req)
	if ip != "192.168.1.100" {
		t.Errorf("Expected IP '192.168.1.100', got '%s'", ip)
	}
}

func TestGetClientIP_XForwardedFor(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	SetTrustedProxyCIDRs([]netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")})
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")

	ip := getClientIP(req)
	if ip != "203.0.113.5" {
		t.Errorf("Expected IP '203.0.113.5', got '%s'", ip)
	}
}

func TestGetClientIP_XRealIP(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	SetTrustedProxyCIDRs([]netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")})
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Real-IP", "203.0.113.10")

	ip := getClientIP(req)
	if ip != "203.0.113.10" {
		t.Errorf("Expected IP '203.0.113.10', got '%s'", ip)
	}
}

func TestGetClientIP_ForwardedIgnoredWhenUntrusted(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	SetTrustedProxyCIDRs(nil)
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.5")
	req.Header.Set("X-Real-IP", "203.0.113.10")

	ip := getClientIP(req)
	if ip != "10.0.0.1" {
		t.Errorf("Expected IP '10.0.0.1', got '%s'", ip)
	}
}

func TestGenerateViewerHash_SecretChange(t *testing.T) {
	if err := SetViewerHashSecret("secret-one", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "test-token-123"
	ip := "192.168.1.1"
	userAgent := "Mozilla/5.0"

	hash1 := GenerateViewerHash(token, ip, userAgent)

	if err := SetViewerHashSecret("secret-two", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	hash2 := GenerateViewerHash(token, ip, userAgent)

	if hash1 == hash2 {
		t.Error("expected different hashes when secret changes")
	}
}

func TestSetViewerHashSecret_EmptyOptional(t *testing.T) {
	if err := SetViewerHashSecret("", false); err != nil {
		t.Fatalf("expected no error for empty optional secret, got %v", err)
	}
	hash := GenerateViewerHash("token", "1.2.3.4", "agent")
	if len(hash) != 64 {
		t.Fatalf("expected hash length 64, got %d", len(hash))
	}
}

func TestSetViewerHashSecret_EmptyRequired(t *testing.T) {
	if err := SetViewerHashSecret("", true); err == nil {
		t.Fatal("expected error for empty required secret")
	}
}

func TestSetViewerHashCookie_WithOptions(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "test-token-abc123xyz"
	viewerHash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	w := httptest.NewRecorder()
	SetViewerHashCookie(w, token, viewerHash, nil, CookieOptions{Secure: true, SameSite: http.SameSiteStrictMode})

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if !cookie.Secure {
		t.Error("Expected Secure cookie")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Errorf("Expected SameSite=Strict, got %v", cookie.SameSite)
	}
}

func TestSetViewerHashCookie_ShortToken(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "short"
	viewerHash := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	w := httptest.NewRecorder()
	SetViewerHashCookie(w, token, viewerHash, nil, CookieOptions{Secure: false, SameSite: http.SameSiteLaxMode})

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Expected 1 cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.Name != "_vh_short" {
		t.Errorf("Expected cookie name '_vh_short', got '%s'", cookie.Name)
	}
	if cookie.Value != viewerHash {
		t.Errorf("Expected cookie value %s, got %s", viewerHash, cookie.Value)
	}
}

func TestGetViewerHash_ShortToken_NoPanic(t *testing.T) {
	if err := SetViewerHashSecret("test-secret", true); err != nil {
		t.Fatalf("set secret: %v", err)
	}
	token := "abc"

	req := httptest.NewRequest("GET", "/s/"+token, nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("User-Agent", "TestAgent/1.0")

	hash := GetViewerHash(req, token)
	if len(hash) != 64 {
		t.Fatalf("expected hash length 64, got %d", len(hash))
	}

	// Ensure short token cookie name is used when present
	existing := "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"
	req = httptest.NewRequest("GET", "/s/"+token, nil)
	req.AddCookie(&http.Cookie{Name: "_vh_short", Value: existing})
	hash = GetViewerHash(req, token)
	if hash != existing {
		t.Fatalf("expected hash from short cookie, got %s", hash)
	}
}
