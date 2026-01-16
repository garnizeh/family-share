package handler

import (
	"database/sql"
	"embed"
	"html/template"
	"net/http"
	"sync"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/storage"
)

type Handler struct {
	db        *sql.DB
	queries   *sqlc.Queries
	storage   *storage.Storage
	templates *template.Template
	embedFS   embed.FS
	tmplMu    sync.RWMutex
}

func New(database *sql.DB, store *storage.Storage, embedFS embed.FS) *Handler {
	// Parse templates from embedded filesystem. Try several common patterns so
	// New works whether the embed.FS contains files under "web/templates/..."
	// (cmd embed) or under "templates/..." (web package embed).
	var tmpl *template.Template
	var err error
	patterns := []string{
		"web/templates/**/*.html",
		"templates/**/*.html",
		"web/templates/*",
		"templates/*",
	}
	for _, p := range patterns {
		tmpl, err = template.ParseFS(embedFS, p)
		if err == nil {
			break
		}
	}
	if err != nil {
		// If no templates found, create an empty template set to avoid panics in tests.
		tmpl = template.New("base")
	}

	return &Handler{
		db:        database,
		queries:   sqlc.New(database),
		storage:   store,
		templates: tmpl,
		embedFS:   embedFS,
	}
}

// RenderTemplate renders a template with data
func (h *Handler) RenderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	h.tmplMu.RLock()
	defer h.tmplMu.RUnlock()

	return h.templates.ExecuteTemplate(w, name, data)
}

// IsHTMX checks if request is an HTMX request
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}
