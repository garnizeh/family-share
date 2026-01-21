package handler

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/middleware"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Health check
	r.Get("/health", h.HealthCheck)

	// Static files (serve from embedded static/ subdirectory)
	if sub, err := fs.Sub(h.embedFS, "static"); err == nil {
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	} else {
		// fall back to the root FS if sub doesn't exist
		r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(h.embedFS))))
	}

	// Serve photo files from data/photos directory
	r.Get("/data/photos/{id}.webp", h.ServePhoto)

	// Public routes
	r.Get("/", h.HomePage)

	// Share link routes
	r.Route("/s", func(r chi.Router) {
		// Will be implemented in task 080
		r.Get("/{token}", h.NotImplementedYet)
	})

	// Admin routes
	r.Route("/admin", func(r chi.Router) {
		// Apply admin auth middleware (stub)
		r.Use(middleware.AdminOnly)
		// Admin pages
		r.Get("/", h.AdminDashboard)
		r.Get("/albums", h.ListAlbums)

		// Album management
		r.Post("/albums", h.CreateAlbum)
		r.Get("/albums/{id}", h.ViewAlbum)
		r.Get("/albums/{id}/edit", h.EditAlbumForm)
		r.Post("/albums/{id}", h.UpdateAlbum)
		r.Put("/albums/{id}", h.UpdateAlbum)
		r.Delete("/albums/{id}", h.DeleteAlbum)

		// Photo upload
		r.Post("/albums/{id}/photos", h.AdminUploadPhotos)

		// Photo management
		r.Delete("/photos/{id}", h.DeletePhoto)
		r.Post("/photos/{id}/set-cover", h.SetCoverPhoto)
	})
}

// Placeholder handlers
func (h *Handler) NotImplementedYet(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.RenderTemplate(w, "home.html", nil); err != nil {
		log.Printf("template render error for home: %v", err)
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	q := sqlc.New(h.db)

	// Get dashboard statistics
	albumCount, _ := q.CountAlbums(r.Context())
	photoCount, _ := q.CountPhotos(r.Context())
	storageBytes, _ := q.GetTotalStorageBytes(r.Context())

	// Convert bytes to MB
	var storageMB float64
	if bytesVal, ok := storageBytes.(int64); ok {
		storageMB = float64(bytesVal) / (1024 * 1024)
	}

	data := struct {
		AlbumCount int64
		PhotoCount int64
		StorageMB  float64
		HasAlbums  bool
	}{
		AlbumCount: albumCount,
		PhotoCount: photoCount,
		StorageMB:  storageMB,
		HasAlbums:  albumCount > 0,
	}

	if err := h.RenderTemplate(w, "admin_dashboard.html", data); err != nil {
		log.Printf("template render error for admin_dashboard: %v", err)
		http.Error(w, "template render error", http.StatusInternalServerError)
	}
}

func (h *Handler) ListAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := h.queries.ListAlbumsWithPhotoCount(r.Context(), sqlc.ListAlbumsWithPhotoCountParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		http.Error(w, "Failed to load albums", http.StatusInternalServerError)
		return
	}

	// Render albums list using the admin layout and a dynamic content fragment
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.RenderTemplate(w, "albums_list.html", albums); err != nil {
		// Log template error for debugging
		log.Printf("template render error for albums_list: %v", err)
		http.Error(w, "template render error", http.StatusInternalServerError)
		return
	}
}
