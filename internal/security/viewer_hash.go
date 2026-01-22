package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// ViewerHashSecret is the secret key for HMAC signing viewer hashes
// In production, load this from environment variable
var ViewerHashSecret = []byte("familyshare-viewer-hash-secret-change-in-production")

// GenerateViewerHash creates a unique hash for a visitor based on token, IP, and User-Agent
func GenerateViewerHash(token, ip, userAgent string) string {
	h := hmac.New(sha256.New, ViewerHashSecret)
	h.Write([]byte(token))
	h.Write([]byte(ip))
	h.Write([]byte(userAgent))
	return hex.EncodeToString(h.Sum(nil))
}

// GetViewerHash retrieves or creates a viewer hash from/to cookie
func GetViewerHash(r *http.Request, token string) string {
	cookieName := "_vh_" + token[:8] // Use first 8 chars of token as cookie prefix

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
func SetViewerHashCookie(w http.ResponseWriter, token, viewerHash string, expiresAt *time.Time) {
	cookieName := "_vh_" + token[:8]

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
		SameSite: http.SameSiteLaxMode,
		Secure:   false, // Set to true in production with HTTPS
	}

	http.SetCookie(w, cookie)
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
