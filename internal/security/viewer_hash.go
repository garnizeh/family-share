package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	viewerHashSecretMu sync.RWMutex
	viewerHashSecret   []byte
)

// SetViewerHashSecret configures the HMAC secret for viewer hashes.
// If secret is empty and required is true, it returns an error.
// If secret is empty and required is false, it logs a warning and uses
// an ephemeral, random secret for this process.
func SetViewerHashSecret(secret string, required bool) error {
	if secret == "" {
		if required {
			return fmt.Errorf("VIEWER_HASH_SECRET is required but not set")
		}
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			return fmt.Errorf("generate viewer hash secret: %w", err)
		}
		viewerHashSecretMu.Lock()
		viewerHashSecret = buf
		viewerHashSecretMu.Unlock()
		log.Printf("warning: VIEWER_HASH_SECRET not set; using ephemeral secret")
		return nil
	}

	viewerHashSecretMu.Lock()
	viewerHashSecret = []byte(secret)
	viewerHashSecretMu.Unlock()
	return nil
}

func getViewerHashSecret() []byte {
	viewerHashSecretMu.RLock()
	defer viewerHashSecretMu.RUnlock()
	return viewerHashSecret
}

// GenerateViewerHash creates a unique hash for a visitor based on token, IP, and User-Agent
func GenerateViewerHash(token, ip, userAgent string) string {
	secret := getViewerHashSecret()
	if len(secret) == 0 {
		_ = SetViewerHashSecret("", false)
		secret = getViewerHashSecret()
	}
	h := hmac.New(sha256.New, secret)
	h.Write([]byte(token))
	h.Write([]byte(ip))
	h.Write([]byte(userAgent))
	return hex.EncodeToString(h.Sum(nil))
}

// GetViewerHash retrieves or creates a viewer hash from/to cookie
func GetViewerHash(r *http.Request, token string) string {
	cookieName := viewerHashCookieName(token)

	// Try to get existing hash from cookie
	cookie, err := r.Cookie(cookieName)
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Generate new hash
	ip := getClientIP(r)
	userAgent := r.UserAgent()
	return GenerateViewerHash(token, ip, userAgent)
}

// SetViewerHashCookie sets the viewer hash cookie
func SetViewerHashCookie(w http.ResponseWriter, token, viewerHash string, expiresAt *time.Time, opts CookieOptions) {
	cookieName := viewerHashCookieName(token)

	// Default expiration: 30 days or share link expiration, whichever is sooner
	maxAge := 30 * 24 * 60 * 60 // 30 days in seconds
	if expiresAt != nil && time.Until(*expiresAt) < 30*24*time.Hour {
		maxAge = int(time.Until(*expiresAt).Seconds())
		if maxAge < 0 {
			maxAge = 0
		}
	}

	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    viewerHash,
		Path:     "/s/" + token,
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: opts.SameSite,
		Secure:   opts.Secure,
	}

	http.SetCookie(w, cookie)
}

func viewerHashCookieName(token string) string {
	if len(token) >= 8 {
		return "_vh_" + token[:8]
	}
	return "_vh_short"
}

// CookieOptions configures cookie security flags.
type CookieOptions struct {
	Secure   bool
	SameSite http.SameSite
}

// ParseSameSite converts a string to an http.SameSite value (default Lax).
func ParseSameSite(value string) http.SameSite {
	switch strings.ToLower(value) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	case "lax":
		fallthrough
	default:
		return http.SameSiteLaxMode
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}
