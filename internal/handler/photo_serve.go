package handler

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"familyshare/internal/storage"

	"github.com/go-chi/chi/v5"
)

// ServePhoto serves a photo file by ID
func (h *Handler) ServePhoto(w http.ResponseWriter, r *http.Request) {
	photoIDStr := chi.URLParam(r, "id")
	photoID, err := strconv.ParseInt(photoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	// Get photo from database to find album_id
	ctx := r.Context()
	photo, err := h.queries.GetPhoto(ctx, photoID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	// Build the file path using storage.PhotoPathAt
	// Extract extension from filename
	ext := strings.ToLower(photo.Format)
	if ext == "" {
		ext = "webp"
	}
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	photoPath := storage.PhotoPathAt(h.storage.BaseDir, photo.AlbumID, photoID, ext, createdAt)

	// Serve the file
	http.ServeFile(w, r, photoPath)
}

// ServeSharedPhoto serves a photo file only when accessed via a valid share token.
func (h *Handler) ServeSharedPhoto(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if token == "" {
		http.NotFound(w, r)
		return
	}

	photoIDStr := chi.URLParam(r, "id")
	photoID, err := strconv.ParseInt(photoIDStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()
	link, err := h.queries.GetShareLinkByToken(ctx, token)
	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		}
		log.Printf("error loading share link for photo: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if link.RevokedAt.Valid {
		http.NotFound(w, r)
		return
	}

	if link.ExpiresAt.Valid && time.Now().UTC().After(link.ExpiresAt.Time) {
		http.NotFound(w, r)
		return
	}

	if link.MaxViews.Valid {
		uniqueViews, err := h.queries.CountUniqueShareLinkViews(ctx, link.ID)
		if err != nil {
			log.Printf("error counting views for shared photo: %v", err)
		} else if uniqueViews >= link.MaxViews.Int64 {
			http.NotFound(w, r)
			return
		}
	}

	photo, err := h.queries.GetPhoto(ctx, photoID)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	switch link.TargetType {
	case "album":
		if photo.AlbumID != link.TargetID {
			http.NotFound(w, r)
			return
		}
	case "photo":
		if photo.ID != link.TargetID {
			http.NotFound(w, r)
			return
		}
	default:
		http.NotFound(w, r)
		return
	}

	ext := strings.ToLower(photo.Format)
	if ext == "" {
		ext = "webp"
	}
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	photoPath := storage.PhotoPathAt(h.storage.BaseDir, photo.AlbumID, photo.ID, ext, createdAt)

	http.ServeFile(w, r, photoPath)
}
