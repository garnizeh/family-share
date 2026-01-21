package handler

import (
	"net/http"
	"strconv"

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

	// Build the file path using storage.PhotoPath
	// Extract extension from filename
	ext := "webp" // default, could parse from photo.Filename if needed
	
	photoPath := storage.PhotoPath(h.storage.BaseDir, photo.AlbumID, photoID, ext)
	
	// Serve the file
	http.ServeFile(w, r, photoPath)
}
