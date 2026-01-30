package handler

import (
	"io/fs"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/metrics"
	"familyshare/internal/middleware"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Health check
	r.Get("/health", h.HealthCheck)

	// Static files (serve from embedded static/ subdirectory)
	// Static files: serve with long-lived immutable cache headers
	if sub, err := fs.Sub(h.embedFS, "static"); err == nil {
		staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(sub)))
		r.Handle("/static/*", cacheWrapper(staticHandler, "public, max-age=31536000, immutable"))
	} else {
		// fall back to the root FS if sub doesn't exist
		staticHandler := http.StripPrefix("/static/", http.FileServer(http.FS(h.embedFS)))
		r.Handle("/static/*", cacheWrapper(staticHandler, "public, max-age=31536000, immutable"))
	}

	// Public routes
	r.Get("/", h.HomePage)

	// Share link routes - apply rate limiting to prevent brute-force token guessing
	r.Route("/s", func(r chi.Router) {
		shareLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
			RequestsPerMinute: h.config.RateLimitShare,
			LockoutDuration:   5 * time.Minute,
			MaxViolations:     10,
			TemplateRenderer:  h,
			TrustedProxyCIDRs: h.config.TrustedProxyCIDRs,
		})
		r.Use(shareLimiter.Middleware())
		r.Get("/{token}", h.ViewShareLink)
		r.Get("/{token}/photos/{id}.webp", h.ServeSharedPhoto)
	})

	// Admin routes - apply stricter rate limiting
	r.Route("/admin", func(r chi.Router) {
		csrf := middleware.NewCSRF(h.config.CSRFSecret)
		r.Use(csrf.Middleware())

		// Apply rate limiting before auth to prevent brute-force login attempts
		adminLimiter := middleware.NewRateLimiter(middleware.RateLimitConfig{
			RequestsPerMinute: h.config.RateLimitAdmin,
			LockoutDuration:   5 * time.Minute,
			MaxViolations:     10,
			TemplateRenderer:  h,
			TrustedProxyCIDRs: h.config.TrustedProxyCIDRs,
		})
		r.Use(adminLimiter.Middleware())

		// Login routes (not protected)
		r.Get("/login", h.LoginPage)
		r.Post("/login", h.Login)

		// Protected admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(h.db))

			// Logout
			r.Post("/logout", h.Logout)

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
			r.Get("/photos/{id}.webp", h.ServePhoto)
			r.Delete("/photos/{id}", h.DeletePhoto)
			r.Post("/photos/{id}/set-cover", h.SetCoverPhoto)

			// Share link management
			r.Get("/shares", h.ListShareLinks)
			r.Post("/shares", h.CreateShareLink)
			r.Delete("/shares/{id}", h.RevokeShareLink)
		})
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

	// Get activity metrics
	stats, err := h.metrics.GetStats(r.Context())
	if err != nil {
		log.Printf("failed to get metrics: %v", err)
		// Continue with empty stats rather than failing the whole page
		stats = &metrics.Stats{}
	}

	data := struct {
		AlbumCount int64
		PhotoCount int64
		StorageMB  float64
		HasAlbums  bool
		Stats      *metrics.Stats
	}{
		AlbumCount: albumCount,
		PhotoCount: photoCount,
		StorageMB:  storageMB,
		HasAlbums:  albumCount > 0,
		Stats:      stats,
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
