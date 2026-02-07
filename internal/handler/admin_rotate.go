package handler

import (
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/pipeline"
	"familyshare/internal/storage"
)

// AdminRotatePhoto handles POST /admin/photos/{id}/rotate
func (h *Handler) AdminRotatePhoto(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 0, 64)
	if err != nil {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	angleStr := r.FormValue("angle")
	angle, err := strconv.Atoi(angleStr)
	if err != nil {
		http.Error(w, "Invalid angle", http.StatusBadRequest)
		return
	}

	// Validate angle (90, -90, 180, 270)
	// We allow 270 as alias for -90
	if angle != 90 && angle != -90 && angle != 180 && angle != 270 {
		http.Error(w, "Angle must be 90, 180, or 270 (-90)", http.StatusBadRequest)
		return
	}

	// Get photo info to find path
	q := sqlc.New(h.db)
	photo, err := q.GetPhoto(r.Context(), id)
	if err != nil {
		http.Error(w, "Photo not found", http.StatusNotFound)
		return
	}

	// Construct full file path
	// Look up createdAt for path resolution
	createdAt := photo.CreatedAt.Time
	if !photo.CreatedAt.Valid {
		// Fallback should not happen in legitimate cases but handle defensively
		// Maybe default to now or zero time? Path resolution relies on it.
		// If data is old/corrupt, this might fail to find the file.
		log.Printf("photo %d has no created_at, using zero time", id)
	}

	photoPath := storage.PhotoPathAt(h.storage.BaseDir, photo.AlbumID, photo.ID, "webp", createdAt)

	// Perform rotation
	// Note: angle in pipeline.Rotate is counter-clockwise.
	// 90 -> Left
	// -90 -> Right
	newWidth, newHeight, newSize, err := pipeline.Rotate(photoPath, angle)
	if err != nil {
		log.Printf("failed to rotate photo %d: %v", id, err)
		http.Error(w, "Failed to process image rotation", http.StatusInternalServerError)
		return
	}

	// Update DB
	err = q.UpdatePhotoDimensions(r.Context(), sqlc.UpdatePhotoDimensionsParams{
		Width:     int64(newWidth),
		Height:    int64(newHeight),
		SizeBytes: newSize,
		ID:        id,
	})
	if err != nil {
		log.Printf("failed to update photo dimensions %d: %v", id, err)
		http.Error(w, "Failed to update database", http.StatusInternalServerError)
		return
	}

	// Fetch updated photo for rendering
	updatedPhoto, err := q.GetPhoto(r.Context(), id)
	if err != nil {
		// Should not happen as we just updated it, but handle nicely
		log.Printf("failed to get updated photo %d: %v", id, err)
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Render the single photo card
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.RenderTemplate(w, "photo_card", updatedPhoto); err != nil {
		log.Printf("template render error: %v", err)
		// Fallback to refresh if template fails
		w.Header().Set("HX-Refresh", "true")
		w.WriteHeader(http.StatusOK)
	}
}
