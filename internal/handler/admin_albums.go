package handler

import (
	"database/sql"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/storage"
)

// CreateAlbum handles POST /admin/albums
func (h *Handler) CreateAlbum(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	title := r.PostFormValue("title")
	desc := r.PostFormValue("description")
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)
	alb, err := q.CreateAlbum(r.Context(), sqlc.CreateAlbumParams{Title: title, Description: sql.NullString{String: desc, Valid: desc != ""}})
	if err != nil {
		http.Error(w, "failed to create album", http.StatusInternalServerError)
		return
	}

	// If HTMX request, redirect to album detail page
	if IsHTMX(r) {
		w.Header().Set("HX-Redirect", "/admin/albums/"+strconv.FormatInt(alb.ID, 10))
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/admin/albums/"+strconv.FormatInt(alb.ID, 10), http.StatusSeeOther)
}

// EditAlbumForm returns the album_form partial prefilled for editing
func (h *Handler) EditAlbumForm(w http.ResponseWriter, r *http.Request) {
	idstr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)
	alb, err := q.GetAlbum(r.Context(), id)
	if err != nil {
		http.Error(w, "album not found", http.StatusNotFound)
		return
	}

	// Check if request is from detail page via query param
	templateName := "album_edit_form.html"
	if r.URL.Query().Get("view") == "detail" {
		templateName = "album_edit_form_detail.html"
	}

	// Render the album edit form partial with album data
	if err := h.RenderTemplate(w, templateName, alb); err != nil {
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

// ViewAlbum handles GET /admin/albums/{id}
func (h *Handler) ViewAlbum(w http.ResponseWriter, r *http.Request) {
	idstr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)
	alb, err := q.GetAlbum(r.Context(), id)
	if err != nil {
		http.Error(w, "album not found", http.StatusNotFound)
		return
	}

	// fetch photos for album
	photos, _ := q.ListPhotosByAlbum(r.Context(), sqlc.ListPhotosByAlbumParams{AlbumID: id, Limit: 100, Offset: 0})

	data := struct {
		Album  sqlc.Album
		Photos []sqlc.Photo
	}{
		Album:  alb,
		Photos: photos,
	}

	if err := h.RenderTemplate(w, "album_detail.html", data); err != nil {
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

// UpdateAlbum handles POST /admin/albums/{id}
func (h *Handler) UpdateAlbum(w http.ResponseWriter, r *http.Request) {
	idstr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}
	title := r.PostFormValue("title")
	desc := r.PostFormValue("description")
	if title == "" {
		http.Error(w, "title required", http.StatusBadRequest)
		return
	}

	q := sqlc.New(h.db)
	err = q.UpdateAlbum(r.Context(), sqlc.UpdateAlbumParams{Title: title, Description: sql.NullString{String: desc, Valid: desc != ""}, CoverPhotoID: sql.NullInt64{}, ID: id})
	if err != nil {
		http.Error(w, "failed to update", http.StatusInternalServerError)
		return
	}

	if IsHTMX(r) {
		alb, _ := q.GetAlbum(r.Context(), id)
		w.Header().Set("HX-Trigger", "closeModal")
		_ = h.RenderTemplate(w, "album_row.html", alb)
		return
	}
	http.Redirect(w, r, "/admin/albums", http.StatusSeeOther)
}

// DeleteAlbum handles DELETE /admin/albums/{id}
func (h *Handler) DeleteAlbum(w http.ResponseWriter, r *http.Request) {
	idstr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	q := sqlc.New(h.db)

	// First, get all photos in this album
	photos, err := q.ListPhotosByAlbum(ctx, sqlc.ListPhotosByAlbumParams{
		AlbumID: id,
		Limit:   1000, // Get all photos
		Offset:  0,
	})
	if err != nil {
		http.Error(w, "failed to list photos", http.StatusInternalServerError)
		return
	}

	// Delete all photo files from disk
	for _, photo := range photos {
		photoPath := storage.PhotoPath(h.storage.BaseDir, photo.AlbumID, photo.ID, "webp")
		// Ignore errors if file doesn't exist
		_ = os.Remove(photoPath)
	}

	// Delete the album (cascade will delete photos from DB via foreign key)
	if err := q.DeleteAlbum(ctx, id); err != nil {
		http.Error(w, "failed to delete", http.StatusInternalServerError)
		return
	}

	// For HTMX delete, return 200 with empty body so HTMX will replace the
	// target element with nothing (effectively removing it from the DOM).
	if IsHTMX(r) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/admin/albums", http.StatusSeeOther)
}
