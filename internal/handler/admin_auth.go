package handler

import (
	"log"
	"net/http"
	"time"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/security"
)

const (
	sessionCookieName = "session_id"
	sessionDuration   = 24 * time.Hour
)

// LoginPage shows the admin login form
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	// If already logged in, redirect to admin dashboard
	if _, err := r.Cookie(sessionCookieName); err == nil {
		// Validate session
		if h.isValidSession(r) {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := struct {
		Error string
	}{
		Error: r.URL.Query().Get("error"),
	}

	if err := h.RenderTemplate(w, "login.html", data); err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

// Login handles admin login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/admin/login?error=invalid_request", http.StatusSeeOther)
		return
	}

	password := r.PostFormValue("password")
	if password == "" {
		http.Redirect(w, r, "/admin/login?error=password_required", http.StatusSeeOther)
		return
	}

	// Verify password against configured hash
	if !security.VerifyPassword(h.config.AdminPasswordHash, password) {
		log.Printf("Failed login attempt from %s", r.RemoteAddr)
		http.Redirect(w, r, "/admin/login?error=invalid_password", http.StatusSeeOther)
		return
	}

	// Create session
	sessionID, err := security.GenerateSecureToken()
	if err != nil {
		log.Printf("Failed to generate session token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	q := sqlc.New(h.db)
	expiresAt := time.Now().UTC().Add(sessionDuration)

	_, err = q.CreateSession(r.Context(), sqlc.CreateSessionParams{
		ID:        sessionID,
		UserID:    "admin", // For MVP, only one admin user
		ExpiresAt: expiresAt,
	})
	if err != nil {
		log.Printf("Failed to create session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	cookieOpts := h.cookieOptions(r)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cookieOpts.Secure,
		SameSite: cookieOpts.SameSite,
	})

	log.Printf("Successful login from %s", r.RemoteAddr)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

// Logout handles admin logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil {
		// Delete session from database
		q := sqlc.New(h.db)
		if err := q.DeleteSession(r.Context(), cookie.Value); err != nil {
			log.Printf("Failed to delete session: %v", err)
		}
	}

	// Clear cookie
	cookieOpts := h.cookieOptions(r)
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   cookieOpts.Secure,
		SameSite: cookieOpts.SameSite,
	})

	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// isValidSession checks if the request has a valid session
func (h *Handler) isValidSession(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}

	q := sqlc.New(h.db)
	session, err := q.GetSession(r.Context(), cookie.Value)
	if err != nil {
		return false
	}

	// Check if session is expired
	if time.Now().UTC().After(session.ExpiresAt) {
		// Clean up expired session
		q.DeleteSession(r.Context(), cookie.Value)
		return false
	}

	return true
}
