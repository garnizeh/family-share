package handler

import (
	"database/sql"
	"familyshare/internal/db/sqlc"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// SetCoverPhoto sets a photo as the album cover
func (h *Handler) SetCoverPhoto(w http.ResponseWriter, r *http.Request) {
	photoIDStr := chi.URLParam(r, "id")
	photoID, err := strconv.ParseInt(photoIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Get the photo to find its album
	photo, err := h.queries.GetPhoto(ctx, photoID)
	if err != nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Update the album's cover photo
	err = h.queries.SetAlbumCover(ctx, sqlc.SetAlbumCoverParams{
		CoverPhotoID: sql.NullInt64{Int64: photoID, Valid: true},
		ID:           photo.AlbumID,
	})
	if err != nil {
		http.Error(w, "Failed to set cover photo", http.StatusInternalServerError)
		return
	}

	// Return success - HTMX will handle UI update
	w.Header().Set("HX-Trigger", "coverPhotoUpdated")
	w.WriteHeader(http.StatusOK)
}
