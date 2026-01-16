package handler

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db/sqlc"
)

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Health check
	r.Get("/health", h.HealthCheck)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.FS(h.embedFS))))

	// Public routes
	r.Get("/", h.HomePage)

	// Share link routes
	r.Route("/s", func(r chi.Router) {
		// Will be implemented in task 080
		r.Get("/{token}", h.NotImplementedYet)
	})

	// Admin routes (authentication middleware will be added in task 090)
	r.Route("/admin", func(r chi.Router) {
		// Admin pages
		r.Get("/", h.AdminDashboard)
		r.Get("/albums", h.ListAlbums)

		// Album management (will be implemented in task 060)
		r.Post("/albums", h.NotImplementedYet)
		r.Get("/albums/{id}", h.NotImplementedYet)

		// Photo upload (will be implemented in task 055)
		r.Post("/albums/{id}/photos", h.NotImplementedYet)
	})
}

// Placeholder handlers
func (h *Handler) NotImplementedYet(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not implemented yet", http.StatusNotImplemented)
}

func (h *Handler) HomePage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>FamilyShare</title></head>
<body>
	<h1>FamilyShare</h1>
	<p>Photo sharing app - Coming soon!</p>
</body>
</html>`))
}

func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Admin Dashboard</title></head>
<body>
	<h1>Admin Dashboard</h1>
	<p>Admin interface - Coming soon!</p>
	<ul>
		<li><a href="/admin/albums">Manage Albums</a></li>
	</ul>
</body>
</html>`))
}

func (h *Handler) ListAlbums(w http.ResponseWriter, r *http.Request) {
	albums, err := h.queries.ListAlbums(r.Context(), sqlc.ListAlbumsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		http.Error(w, "Failed to load albums", http.StatusInternalServerError)
		return
	}

	// Simple HTML output for now (proper templates in task 060)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Albums</title></head>
<body>
	<h1>Albums</h1>
	<p>Albums loaded: ` + fmt.Sprintf("%d", len(albums)) + `</p>
	<a href="/admin">Back to Dashboard</a>
</body>
</html>`))
}
