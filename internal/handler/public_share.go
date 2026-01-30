package handler

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/security"

	"github.com/go-chi/chi/v5"
)

// ViewShareLink handles public access to shared albums or photos via token
func (h *Handler) ViewShareLink(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		h.renderShareExpired(w, "Invalid share link", http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)

	// 1. Load share link
	link, err := q.GetShareLinkByToken(r.Context(), token)
	if err != nil {
		if err == sql.ErrNoRows {
			h.renderShareExpired(w, "Share link not found", http.StatusNotFound)
		} else {
			log.Printf("error loading share link: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// 2. Check if revoked
	if link.RevokedAt.Valid {
		h.renderShareExpired(w, "This share link has been revoked", http.StatusGone)
		return
	}

	// 3. Check expiration
	if link.ExpiresAt.Valid && time.Now().UTC().After(link.ExpiresAt.Time) {
		h.renderShareExpired(w, "This share link has expired", http.StatusGone)
		return
	}

	// 4. Get or create viewer hash
	viewerHash := security.GetViewerHash(r, token)

	// 5. Check view limit (before tracking the view)
	if link.MaxViews.Valid {
		uniqueViews, err := q.CountUniqueShareLinkViews(r.Context(), link.ID)
		if err != nil {
			log.Printf("error counting views: %v", err)
			// Continue anyway, don't block access on count error
		} else if uniqueViews >= link.MaxViews.Int64 {
			h.renderShareExpired(w, "This share link has reached its view limit", http.StatusGone)
			return
		}
	}

	// 6. Track view (INSERT OR IGNORE makes this idempotent)
	err = q.IncrementShareLinkView(r.Context(), sqlc.IncrementShareLinkViewParams{
		ShareLinkID: link.ID,
		ViewerHash:  viewerHash,
	})
	if err != nil {
		log.Printf("error tracking view: %v", err)
		// Continue anyway, tracking is best-effort
	}

	// Log share view event (fire and forget)
	go func(shareID int64) {
		logCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = h.metrics.LogShareView(logCtx, shareID)
	}(link.ID)

	// 7. Set viewer hash cookie for future visits
	security.SetViewerHashCookie(w, token, viewerHash, &link.ExpiresAt.Time, h.cookieOptions(r))

	// 8. Render content based on target type
	switch link.TargetType {
	case "album":
		h.renderShareAlbum(w, r, link)
	case "photo":
		h.renderSharePhoto(w, r, link)
	default:
		h.renderShareExpired(w, "Invalid share link type", http.StatusBadRequest)
	}
}

// renderShareAlbum renders the public album view with HTMX pagination
func (h *Handler) renderShareAlbum(w http.ResponseWriter, r *http.Request, link sqlc.ShareLink) {
	q := sqlc.New(h.db)

	// Load album
	album, err := q.GetAlbum(r.Context(), link.TargetID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.renderShareExpired(w, "Album not found", http.StatusNotFound)
		} else {
			log.Printf("error loading album: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Get pagination parameters
	pageNum := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}

	const pageSize = 20
	offset := (pageNum - 1) * pageSize

	// Load photos for this page (fetch one extra to check if there are more)
	photos, err := q.ListPhotosByAlbum(r.Context(), sqlc.ListPhotosByAlbumParams{
		AlbumID: album.ID,
		Limit:   int64(pageSize + 1),
		Offset:  int64(offset),
	})
	if err != nil {
		log.Printf("error loading photos: %v", err)
		photos = []sqlc.Photo{} // Show empty album on error
	}

	// Check if there are more pages
	hasMore := len(photos) > pageSize
	if hasMore {
		photos = photos[:pageSize] // Trim the extra photo
	}

	// Check if this is an HTMX request
	isHTMX := r.Header.Get("HX-Request") == "true"

	data := struct {
		Album    sqlc.Album
		Photos   []sqlc.Photo
		Token    string
		Page     int
		NextPage int
		HasMore  bool
	}{
		Album:    album,
		Photos:   photos,
		Token:    link.Token,
		Page:     pageNum,
		NextPage: pageNum + 1,
		HasMore:  hasMore,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var templateName string
	if isHTMX {
		templateName = "photo_grid_partial.html"
	} else {
		templateName = "share_album.html"
	}

	if err := h.RenderTemplate(w, templateName, data); err != nil {
		log.Printf("template render error for %s: %v", templateName, err)
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

// renderSharePhoto renders the public single photo view
func (h *Handler) renderSharePhoto(w http.ResponseWriter, r *http.Request, link sqlc.ShareLink) {
	q := sqlc.New(h.db)

	// Load photo
	photo, err := q.GetPhoto(r.Context(), link.TargetID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.renderShareExpired(w, "Photo not found", http.StatusNotFound)
		} else {
			log.Printf("error loading photo: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Load album info
	album, err := q.GetAlbum(r.Context(), photo.AlbumID)
	if err != nil {
		log.Printf("error loading album for photo: %v", err)
		// Continue with empty album
	}

	data := struct {
		Photo sqlc.Photo
		Album sqlc.Album
		Token string
	}{
		Photo: photo,
		Album: album,
		Token: link.Token,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.RenderTemplate(w, "share_photo.html", data); err != nil {
		log.Printf("template render error for share_photo: %v", err)
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

// renderShareExpired renders the error page for expired/invalid links
func (h *Handler) renderShareExpired(w http.ResponseWriter, message string, statusCode int) {
	data := struct {
		Message string
	}{
		Message: message,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := h.RenderTemplate(w, "share_expired.html", data); err != nil {
		log.Printf("template render error for share_expired: %v", err)
		http.Error(w, fmt.Sprintf("Error: %s", message), statusCode)
	}
}
