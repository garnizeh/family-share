package handler

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/security"
)

// ListShareLinks handles GET /admin/shares
func (h *Handler) ListShareLinks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	q := sqlc.New(h.db)

	// Check if user wants to see revoked shares
	showRevoked := r.URL.Query().Get("show_revoked") == "true"

	// Get all share links with details
	allShares, err := q.ListShareLinksWithDetails(r.Context(), sqlc.ListShareLinksWithDetailsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		log.Printf("failed to list share links: %v", err)
		http.Error(w, "failed to list share links", http.StatusInternalServerError)
		return
	}

	// Filter out revoked shares unless explicitly requested
	var shares []sqlc.ListShareLinksWithDetailsRow
	if showRevoked {
		shares = allShares
	} else {
		for _, share := range allShares {
			if !share.RevokedAt.Valid {
				shares = append(shares, share)
			}
		}
	}

	// Get albums and photos for the form dropdown
	albums, _ := q.ListAlbums(r.Context(), sqlc.ListAlbumsParams{Limit: 100, Offset: 0})
	photos, _ := q.ListAllPhotosWithAlbum(r.Context(), sqlc.ListAllPhotosWithAlbumParams{Limit: 100, Offset: 0})

	data := struct {
		Shares      []sqlc.ListShareLinksWithDetailsRow
		Albums      []sqlc.Album
		Photos      []sqlc.ListAllPhotosWithAlbumRow
		BaseURL     string
		ShowRevoked bool
	}{
		Shares:      shares,
		Albums:      albums,
		Photos:      photos,
		BaseURL:     getBaseURL(r),
		ShowRevoked: showRevoked,
	}

	if err := h.RenderTemplate(w, "shares_list.html", data); err != nil {
		log.Printf("template render error: %v", err)
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

// CreateShareLink handles POST /admin/shares
func (h *Handler) CreateShareLink(w http.ResponseWriter, r *http.Request) {
	log.Printf("CreateShareLink called: method=%s, content-type=%s", r.Method, r.Header.Get("Content-Type"))

	if err := r.ParseForm(); err != nil {
		log.Printf("failed to parse form: %v", err)
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	targetType := r.PostFormValue("target_type")
	targetIDStr := r.PostFormValue("target_id")
	maxViewsStr := r.PostFormValue("max_views")
	expiresAtStr := r.PostFormValue("expires_at")

	log.Printf("Form values: target_type=%s, target_id=%s, max_views=%s, expires_at=%s",
		targetType, targetIDStr, maxViewsStr, expiresAtStr)

	// Validate target type
	if targetType != "album" && targetType != "photo" {
		log.Printf("invalid target_type: %s", targetType)
		http.Error(w, "invalid target_type", http.StatusBadRequest)
		return
	}

	// Parse target ID
	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil || targetID <= 0 {
		http.Error(w, "invalid target_id", http.StatusBadRequest)
		return
	}

	// Parse max views (optional)
	var maxViews sql.NullInt64
	if maxViewsStr != "" {
		mv, err := strconv.ParseInt(maxViewsStr, 10, 64)
		if err != nil || mv <= 0 {
			http.Error(w, "invalid max_views", http.StatusBadRequest)
			return
		}
		maxViews = sql.NullInt64{Int64: mv, Valid: true}
	}

	// Parse expires_at (optional)
	var expiresAt sql.NullTime
	if expiresAtStr != "" {
		// Parse in UTC timezone
		t, err := time.ParseInLocation("2006-01-02T15:04", expiresAtStr, time.UTC)
		if err != nil {
			http.Error(w, "invalid expires_at format", http.StatusBadRequest)
			return
		}
		expiresAt = sql.NullTime{Time: t, Valid: true}
	}

	// Parse message (optional)
	message := r.PostFormValue("message")
	var messageSQL sql.NullString
	if message != "" {
		messageSQL = sql.NullString{String: message, Valid: true}
	}

	q := sqlc.New(h.db)

	// Verify target exists
	switch targetType {
	case "album":
		if _, err := q.GetAlbum(r.Context(), targetID); err != nil {
			http.Error(w, "album not found", http.StatusNotFound)
			return
		}
	case "photo":
		if _, err := q.GetPhoto(r.Context(), targetID); err != nil {
			http.Error(w, "photo not found", http.StatusNotFound)
			return
		}
	}

	// Generate secure token with retry logic for uniqueness
	var token string
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		token, err = security.GenerateSecureToken()
		if err != nil {
			log.Printf("failed to generate token: %v", err)
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		// Try to create share link
		share, err := q.CreateShareLink(r.Context(), sqlc.CreateShareLinkParams{
			Token:      token,
			TargetType: targetType,
			TargetID:   targetID,
			MaxViews:   maxViews,
			ExpiresAt:  expiresAt,
			Message:    messageSQL,
		})

		if err == nil {
			// Success - return the share link row
			data := struct {
				Share   sqlc.ShareLink
				BaseURL string
			}{
				Share:   share,
				BaseURL: getBaseURL(r),
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if err := h.RenderTemplate(w, "share_row.html", data); err != nil {
				log.Printf("template render error: %v", err)
				http.Error(w, "template render error", http.StatusInternalServerError)
			}
			return
		}

		// Check if error is due to unique constraint violation
		// SQLite error for unique constraint is "UNIQUE constraint failed"
		if i < maxRetries-1 {
			log.Printf("token collision, retrying (%d/%d): %v", i+1, maxRetries, err)
			continue
		}

		// Max retries exceeded
		log.Printf("failed to create share link after %d retries: %v", maxRetries, err)
		http.Error(w, "failed to create share link", http.StatusInternalServerError)
		return
	}
}

// RevokeShareLink handles DELETE /admin/shares/{id}
func (h *Handler) RevokeShareLink(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)

	// Verify share link exists
	if _, err := q.GetShareLink(r.Context(), id); err != nil {
		http.Error(w, "share link not found", http.StatusNotFound)
		return
	}

	// Revoke the share link
	if err := q.RevokeShareLink(r.Context(), id); err != nil {
		log.Printf("failed to revoke share link: %v", err)
		http.Error(w, "failed to revoke share link", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// getBaseURL extracts the base URL from the request
func getBaseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	// Check X-Forwarded-Proto header (common in reverse proxy setups)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	return scheme + "://" + r.Host
}
