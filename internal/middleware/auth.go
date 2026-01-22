package middleware

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"familyshare/internal/db/sqlc"
)

const sessionCookieName = "session_id"

// SessionValidator interface for validating sessions
type SessionValidator interface {
	GetSession(ctx context.Context, id string) (sqlc.Session, error)
	DeleteSession(ctx context.Context, id string) error
}

// RequireAuth is middleware that requires a valid session
func RequireAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			q := sqlc.New(db)
			session, err := q.GetSession(r.Context(), cookie.Value)
			if err != nil {
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			// Check if session is expired
			if time.Now().UTC().After(session.ExpiresAt) {
				// Clean up expired session
				q.DeleteSession(r.Context(), cookie.Value)
				http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
				return
			}

			// Session is valid, continue
			next.ServeHTTP(w, r)
		})
	}
}

// AdminOnly is a stub middleware for admin authentication.
// DEPRECATED: Use RequireAuth instead
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Stub: allow all requests but could check cookies/headers here.
		next.ServeHTTP(w, r)
	})
}
