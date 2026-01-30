package middleware

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	csrfFormField  = "csrf_token"
)

type CSRF struct {
	secret     []byte
	cookieName string
	headerName string
	formField  string
	maxAge     time.Duration
}

// NewCSRF creates a CSRF middleware using a HMAC-signed token stored in a cookie.
func NewCSRF(secret string) *CSRF {
	secretBytes := []byte(secret)
	if len(secretBytes) == 0 {
		secretBytes = make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			log.Printf("csrf: failed to generate secret, using fallback: %v", err)
			secretBytes = []byte("fallback-insecure-secret")
		} else {
			log.Printf("csrf: CSRF_SECRET not set, generated ephemeral secret")
		}
	}

	return &CSRF{
		secret:     secretBytes,
		cookieName: csrfCookieName,
		headerName: csrfHeaderName,
		formField:  csrfFormField,
		maxAge:     24 * time.Hour,
	}
}

func (c *CSRF) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenCookie, err := r.Cookie(c.cookieName)
			cookieToken := ""
			if err == nil {
				cookieToken = tokenCookie.Value
			}

			if cookieToken == "" || !c.validateSignedToken(cookieToken) {
				cookieToken = c.newSignedToken()
				c.setCookie(w, r, cookieToken)
			}

			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			requestToken := r.Header.Get(c.headerName)
			if requestToken == "" {
				requestToken = r.FormValue(c.formField)
			}
			if requestToken == "" {
				http.Error(w, "CSRF token missing", http.StatusForbidden)
				return
			}

			if requestToken != cookieToken {
				http.Error(w, "CSRF token invalid", http.StatusForbidden)
				return
			}

			if !c.validateSignedToken(requestToken) {
				http.Error(w, "CSRF token invalid", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (c *CSRF) setCookie(w http.ResponseWriter, r *http.Request, value string) {
	cookie := &http.Cookie{
		Name:     c.cookieName,
		Value:    value,
		Path:     "/admin",
		MaxAge:   int(c.maxAge.Seconds()),
		SameSite: http.SameSiteLaxMode,
		Secure:   r.TLS != nil,
		HttpOnly: false,
	}
	http.SetCookie(w, cookie)
}

func (c *CSRF) newSignedToken() string {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		raw = []byte(time.Now().UTC().String())
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	sig := c.sign(token)
	return token + "." + sig
}

func (c *CSRF) validateSignedToken(value string) bool {
	token, sig, err := splitToken(value)
	if err != nil {
		return false
	}
	expected := c.sign(token)
	return hmac.Equal([]byte(sig), []byte(expected))
}

func (c *CSRF) sign(token string) string {
	mac := hmac.New(sha256.New, c.secret)
	mac.Write([]byte(token))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func splitToken(value string) (string, string, error) {
	parts := strings.SplitN(value, ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("invalid token format")
	}
	return parts[0], parts[1], nil
}
