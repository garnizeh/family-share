package handler

import (
	"database/sql"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"sync"

	"strings"

	"familyshare/internal/config"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/metrics"
	"familyshare/internal/security"
	"familyshare/internal/storage"
)

type Handler struct {
	db        *sql.DB
	queries   *sqlc.Queries
	storage   *storage.Storage
	templates *template.Template
	embedFS   embed.FS
	tmplMu    sync.RWMutex
	config    *config.Config
	metrics   *metrics.Logger
}

func New(database *sql.DB, store *storage.Storage, embedFS embed.FS, cfg *config.Config) *Handler {
	if cfg != nil {
		if err := security.SetViewerHashSecret(cfg.ViewerHashSecret, cfg.RequireViewerHashSecret); err != nil {
			log.Fatalf("viewer hash secret configuration error: %v", err)
		}
	}
	// Parse templates from embedded filesystem. Try several common patterns so
	// New works whether the embed.FS contains files under "web/templates/..."
	// (cmd embed) or under "templates/..." (web package embed).
	var tmpl *template.Template
	var err error

	// Collect all embedded .html files by walking the embed FS so we don't
	// depend on specific glob support or relative path patterns.
	var files []string
	_ = fs.WalkDir(embedFS, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".html") {
			files = append(files, path)
		}
		return nil
	})

	if len(files) == 0 {
		log.Printf("no embedded templates found (walk returned 0 files)")
		tmpl = template.New("base")
	} else {
		log.Printf("template files to parse: %v", files)
		tmpl, err = template.ParseFS(embedFS, files...)
		if err != nil {
			log.Printf("template parse error: %v", err)
			// If parsing fails, fall back to an empty template set to avoid panics in tests.
			tmpl = template.New("base")
		}
	}

	// Log loaded template names for debugging template lookup issues
	var names []string
	for _, t := range tmpl.Templates() {
		if t.Name() != "" {
			names = append(names, t.Name())
		}
	}
	log.Printf("loaded templates: %v", names)

	return &Handler{
		db:        database,
		queries:   sqlc.New(database),
		storage:   store,
		templates: tmpl,
		embedFS:   embedFS,
		config:    cfg,
		metrics:   metrics.New(database),
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

func (h *Handler) cookieOptions(r *http.Request) security.CookieOptions {
	secure := false
	if h.config != nil {
		secure = h.config.ForceHTTPS
		if !secure {
			secure = r.TLS != nil
		}
		return security.CookieOptions{
			Secure:   secure,
			SameSite: security.ParseSameSite(h.config.CookieSameSite),
		}
	}
	if r.TLS != nil {
		secure = true
	}
	return security.CookieOptions{
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}
