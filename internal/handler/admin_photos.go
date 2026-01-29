package handler

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/storage"
)

// DeletePhoto handles DELETE /admin/photos/{id}
func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	idstr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)

	// Get photo details before deleting
	photo, err := q.GetPhoto(r.Context(), id)
	if err != nil {
		http.Error(w, "photo not found", http.StatusNotFound)
		return
	}

	// Clear album cover if this photo is the cover
	if err := q.ClearAlbumCoverIfPhoto(r.Context(), sql.NullInt64{Int64: id, Valid: true}); err != nil {
		log.Printf("failed to clear album cover: %v", err)
		// Continue with deletion even if this fails
	}

	// Delete photo from database
	if err := q.DeletePhoto(r.Context(), id); err != nil {
		log.Printf("failed to delete photo from database: %v", err)
		http.Error(w, "failed to delete photo", http.StatusInternalServerError)
		return
	}

	// Delete photo file from disk using the hierarchical path
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	photoPath := storage.PhotoPathAt(h.storage.BaseDir, photo.AlbumID, id, photo.Format, createdAt)
	if err := os.Remove(photoPath); err != nil {
		log.Printf("failed to delete photo file %s: %v", photoPath, err)
		// Don't return error - photo already deleted from DB
	}

	w.WriteHeader(http.StatusNoContent)
}
