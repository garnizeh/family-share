package handler_test

import (
	"context"
	"encoding/json"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"familyshare/internal/config"
	"familyshare/internal/db"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/web"
)

func TestHealthCheck_OK(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New("./testdata")
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body.Status != "healthy" {
		t.Fatalf("expected healthy status, got %s", body.Status)
	}
}

func TestHealthCheck_DBDown(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}

	store := storage.New("./testdata")
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	// Close DB to simulate outage
	dbConn.Close()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	h.HealthCheck(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	if body.Status != "unhealthy" {
		t.Fatalf("expected unhealthy status, got %s", body.Status)
	}
}

func TestStaticFileServing_and_Routes(t *testing.T) {
	// Determine where the embedded static files live (try common locations)
	var subfs fs.FS
	var err error
	tries := []string{"static", "web/static"}
	for _, p := range tries {
		subfs, err = fs.Sub(web.EmbedFS, p)
		if err == nil {
			break
		}
	}
	if err != nil {
		t.Fatalf("could not locate embedded static dir: %v", err)
	}

	fsrv := http.FileServer(http.FS(subfs))
	ts := httptest.NewServer(fsrv)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/styles.css")
	if err != nil {
		t.Fatalf("get static: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for static file, got %d", res.StatusCode)
	}
}

func TestGracefulShutdown(t *testing.T) {
	// Start a real server on an ephemeral port, then call Shutdown and ensure it returns quickly
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	store := storage.New("./testdata")
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	// Setup chi router
	// Use a real http.Server to test Shutdown behavior
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HomePage(w, r)
	})}

	go func() {
		_ = srv.Serve(ln)
	}()

	// give server a moment
	time.Sleep(50 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}
}
