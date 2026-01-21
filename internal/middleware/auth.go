package middleware

import (
	"net/http"
)

// AdminOnly is a stub middleware for admin authentication.
// For now it simply marks the request as coming from an admin and calls next.
// TODO: Replace with real authentication in a later task.
func AdminOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Stub: allow all requests but could check cookies/headers here.
		next.ServeHTTP(w, r)
	})
}
