package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"familyshare/internal/config"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"

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
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

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
	fw.Write([]byte("notreallyjpeg"))
	mw.Close()

	req := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()
	h.AdminUploadPhotos(w, req)

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
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

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
	io.WriteString(fw, "this is not an image")
	mw.Close()

	req := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()
	h.AdminUploadPhotos(w, req)

	files, err := filepath.Glob(filepath.Join(tmpd, "upload-*.tmp"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("expected no temp files after failure, found: %v", files)
	}
}
