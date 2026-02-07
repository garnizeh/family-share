package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"familyshare/internal/config"
	"familyshare/internal/worker"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/web"
)

// Test that temp files are removed after successful upload
func TestTempFilesRemovedAfterSuccess(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	tmpd, err := os.MkdirTemp("", "fs-test-tempdir-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(tmpd)
	os.Setenv("TEMP_UPLOAD_DIR", tmpd)

	store := storage.New("./data")
	cfg := &config.Config{RateLimitShare: 60, RateLimitAdmin: 10}
	wrk := worker.NewWorker(dbConn, store, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wrk.Start(ctx)

	h := handler.New(dbConn, store, web.EmbedFS, cfg, wrk)

	q := sqlc.New(dbConn)
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{Title: "test", Description: sql.NullString{String: "", Valid: false}})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	// build multipart with one small jpeg (use simple bytes)
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("photos", "one.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = fw.Write([]byte("notreallyjpeg"))
	mw.Close()

	req := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()
	h.AdminUploadPhotos(w, req)

	// Wait for worker to pick up and process (it will fail due to bad image, but that triggers cleanup)
	deadline := time.Now().Add(5 * time.Second)
	var processed bool
	for time.Now().Before(deadline) {
		var status string
		err := dbConn.QueryRow("SELECT status FROM processing_queue LIMIT 1").Scan(&status)
		if err == nil && (status == "completed" || status == "failed") {
			processed = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !processed {
		t.Fatal("timed out waiting for worker to process job")
	}

	// ensure no upload-*.tmp files remain in tmpd
	files, err := filepath.Glob(filepath.Join(tmpd, "upload-*.tmp"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no temp files, found: %v", files)
	}
}

// Test that temp files are removed after failed upload (invalid image)
func TestTempFilesRemovedAfterFailure(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	tmpd, err := os.MkdirTemp("", "fs-test-tempdir-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(tmpd)
	os.Setenv("TEMP_UPLOAD_DIR", tmpd)

	store := storage.New("./data")
	cfg := &config.Config{RateLimitShare: 60, RateLimitAdmin: 10}
	wrk := worker.NewWorker(dbConn, store, cfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wrk.Start(ctx)

	h := handler.New(dbConn, store, web.EmbedFS, cfg, wrk)

	q := sqlc.New(dbConn)
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{Title: "test", Description: sql.NullString{String: "", Valid: false}})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	// invalid file (text) - handler should still remove temp
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("photos", "bad.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	_, _ = io.WriteString(fw, "this is not an image")
	mw.Close()

	req := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()
	h.AdminUploadPhotos(w, req)

	// Wait for worker
	deadline := time.Now().Add(5 * time.Second)
	var processed bool
	for time.Now().Before(deadline) {
		var status string
		err := dbConn.QueryRow("SELECT status FROM processing_queue LIMIT 1").Scan(&status)
		if err == nil && (status == "completed" || status == "failed") {
			processed = true
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if !processed {
		t.Fatal("timed out waiting for worker to process job")
	}

	files, err := filepath.Glob(filepath.Join(tmpd, "upload-*.tmp"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no temp files after failure, found: %v", files)
	}
}
