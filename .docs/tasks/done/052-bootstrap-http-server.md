# Task 052: Bootstrap HTTP Server with Routing and Handler Infrastructure

**Milestone:** Server Foundation  
**Points:** 3 (10 hours)  
**Dependencies:** 050  
**Branch:** `feat/http-server-bootstrap`  
**Labels:** `server`, `routing`, `infrastructure`, `handlers`

## Description
Bootstrap the HTTP server with routing infrastructure, template rendering, and handler framework. This establishes the foundation for all admin and public endpoints before implementing specific handlers like the upload endpoint (task 055).

## Acceptance Criteria
- [ ] Chi router configured with middleware stack
- [ ] Template engine initialized with embedded templates
- [ ] Base handler struct with dependencies (DB, storage, templates)
- [ ] Health check endpoint (`GET /health`)
- [ ] Static file serving configured (`/static/*`)
- [ ] Basic route groups for `/admin` and `/s` (share links)
- [ ] Graceful shutdown on SIGTERM/SIGINT
- [ ] Server starts successfully and listens on configured port
- [ ] Environment-based configuration (port, data dir, db path)

## Files to Add/Modify
- `cmd/app/main.go` — server bootstrap and initialization
- `internal/handler/handler.go` — base handler struct and constructor
- `internal/handler/health.go` — health check endpoint
- `internal/handler/routes.go` — route registration
- `internal/config/config.go` — environment-based config
- `web/templates/layout/base.html` — base HTML layout template
- `web/templates/admin/layout.html` — admin layout template
- `web/static/styles.css` — placeholder CSS file

## Server Bootstrap Logic (main.go)

```go
package main

import (
	"context"
	"embed"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"family-share/internal/config"
	"family-share/internal/db"
	"family-share/internal/handler"
	"family-share/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed web/templates/* web/static/*
var embedFS embed.FS

func main() {
	// Load config from environment
	cfg := config.Load()
	
	// Initialize database
	database, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()
	
	// Run migrations
	if err := db.RunMigrations(database); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	
	// Initialize storage
	store := storage.New(cfg.DataDir)
	
	// Initialize router
	r := chi.NewRouter()
	
	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	
	// Initialize handlers
	h := handler.New(database, store, embedFS)
	
	// Register routes
	h.RegisterRoutes(r)
	
	// Create server
	srv := &http.Server{
		Addr:         cfg.ServerAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	// Graceful shutdown
	go func() {
		log.Printf("FamilyShare starting on %s", cfg.ServerAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()
	
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	
	log.Println("Server exited")
}
```

## Handler Infrastructure (handler.go)

```go
package handler

import (
	"database/sql"
	"embed"
	"html/template"
	"sync"

	"family-share/internal/db/sqlc"
	"family-share/internal/storage"
)

type Handler struct {
	db        *sql.DB
	queries   *sqlc.Queries
	storage   *storage.Storage
	templates *template.Template
	tmplMu    sync.RWMutex
}

func New(database *sql.DB, store *storage.Storage, embedFS embed.FS) *Handler {
	// Parse templates from embedded filesystem
	tmpl := template.Must(template.ParseFS(embedFS, "web/templates/**/*.html"))
	
	return &Handler{
		db:        database,
		queries:   sqlc.New(database),
		storage:   store,
		templates: tmpl,
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
```

## Route Registration (routes.go)

```go
package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
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
```

## Configuration (config.go)

```go
package config

import "os"

type Config struct {
	ServerAddr   string
	DatabasePath string
	DataDir      string
}

func Load() *Config {
	return &Config{
		ServerAddr:   getEnv("SERVER_ADDR", ":8080"),
		DatabasePath: getEnv("DATABASE_PATH", "./data/familyshare.db"),
		DataDir:      getEnv("DATA_DIR", "./data"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
```

## Health Check (health.go)

```go
package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	// Ping database to verify connection
	ctx := r.Context()
	if err := h.db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(HealthResponse{
			Status:    "unhealthy",
			Timestamp: time.Now(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
	})
}
```

## Base Templates

### web/templates/layout/base.html
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}FamilyShare{{end}}</title>
    <link rel="stylesheet" href="/static/styles.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
    <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
</head>
<body>
    {{block "content" .}}{{end}}
</body>
</html>
```

### web/static/styles.css
```css
/* Minimal reset and base styles */
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

body {
    font-family: system-ui, -apple-system, sans-serif;
    line-height: 1.5;
    color: #333;
    background: #f5f5f5;
}

/* Placeholder - TailwindCSS will be added later */
```

## Dependencies to Add

```bash
go get github.com/go-chi/chi/v5
```

## Tests Required
- [x] Unit test: config loading with environment variables
- [x] Unit test: config loading with defaults
- [x] Integration test: server starts and responds to /health
- [x] Integration test: static files are served correctly
- [x] Integration test: graceful shutdown works
- [x] Integration test: routes are registered correctly

## Test Example (handler_test.go)

```go
package handler_test

import (
	"database/sql"
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"family-share/internal/handler"
	"family-share/internal/storage"
)

//go:embed testdata/templates/*
var testEmbedFS embed.FS

func TestHealthCheck(t *testing.T) {
	// Setup test database
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	
	store := storage.New("./testdata")
	h := handler.New(db, store, testEmbedFS)
	
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	
	h.HealthCheck(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	
	// Verify JSON response
	// ...
}
```

## Manual Testing Steps

```bash
# Build and run server
go build -o bin/familyshare ./cmd/app
./bin/familyshare

# In another terminal:
# Test health check
curl http://localhost:8080/health

# Expected: {"status":"healthy","timestamp":"..."}

# Test routes
curl http://localhost:8080/
curl http://localhost:8080/admin
curl http://localhost:8080/admin/albums

# Test graceful shutdown
# Press Ctrl+C in server terminal
# Should see: "Shutting down server..." then "Server exited"
```

## PR Checklist
- [x] Chi router configured with standard middleware
- [x] Database initialized and migrations run on startup
- [x] Handler struct properly initialized with dependencies
- [x] Health check returns JSON with database ping
- [x] Graceful shutdown handles SIGTERM and SIGINT
- [x] Static files served from embedded filesystem
- [x] Route groups established for admin and public areas
- [x] Placeholder handlers return 501 Not Implemented
- [x] Configuration loads from environment with sensible defaults
- [x] Tests pass: `go test ./internal/handler/... -v`
- [x] Tests pass: `go test ./internal/config/... -v`
- [x] Server starts without errors
- [x] Manual smoke tests pass (health, routes, shutdown)

## Git Workflow
```bash
git checkout -b feat/http-server-bootstrap

# Add chi dependency
go get github.com/go-chi/chi/v5

# Implement server bootstrap
# Create handler infrastructure
# Add configuration
# Add health check
# Add route registration
# Add base templates

# Test
go test ./... -v
go build -o bin/familyshare ./cmd/app

# Manual testing...

git add .
git commit -m "feat: bootstrap HTTP server with chi routing and handler infrastructure"
git push origin feat/http-server-bootstrap
# Open PR: "Bootstrap HTTP server with routing and handler infrastructure"
```

## Notes
- Keep handlers simple - just establish routing patterns
- Authentication middleware will be added in task 090
- Actual admin UI templates will be implemented in tasks 060, 065, 075
- For MVP, admin auth is not required (can be added later)
- Use embedded filesystem for templates and static files (production-ready)
- Ensure context is passed through all handler methods
- Log startup configuration for debugging
- Consider adding request logging middleware for visibility

## Post-Completion Validation
After this task, the following should work:
1. Server starts on configured port
2. Health check endpoint responds with JSON
3. Database is initialized and accessible
4. Static files are served
5. Routes return appropriate responses (even if placeholder)
6. Graceful shutdown works cleanly
7. Ready for task 055 (upload handler) implementation
