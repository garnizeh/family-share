package handler_test

import (
	"bytes"
	"context"
	"database/sql"
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

// Test that a temp file is created while the upload is being streamed.
func TestTempFileCreatedDuringUpload(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	tmpd, err := os.MkdirTemp("", "fs-stream-tempdir-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(tmpd)
	os.Setenv("TEMP_UPLOAD_DIR", tmpd)

	store := storage.New("./data")
	h := handler.New(dbConn, store, web.EmbedFS)

	q := sqlc.New(dbConn)
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{Title: "test", Description: sql.NullString{String: "", Valid: false}})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	// Writer goroutine: write a part slowly
	go func() {
		defer pw.Close()
		fw, err := mw.CreateFormFile("photos", "stream.jpg")
		if err != nil {
			return
		}
		// write initial chunk
		fw.Write(bytes.Repeat([]byte("a"), 1024))
		// flush by sleeping to allow handler to create temp file
		time.Sleep(200 * time.Millisecond)
		// finish
		fw.Write(bytes.Repeat([]byte("b"), 1024*2))
		mw.Close()
	}()

	req := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", pr)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		h.AdminUploadPhotos(w, req)
		close(done)
	}()

	// Poll tmp dir for up to 1s to see upload-*.tmp created
	var found string
	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		files, _ := filepath.Glob(filepath.Join(tmpd, "upload-*.tmp"))
		if len(files) > 0 {
			found = files[0]
			break
		}
		time.Sleep(20 * time.Millisecond)
	}

	if found == "" {
		t.Fatalf("expected temp file to be created during streaming, none found in %s", tmpd)
	}

	// Wait for handler to finish
	<-done

	// After completion, temp file should be removed
	if _, err := os.Stat(found); !os.IsNotExist(err) {
		t.Fatalf("expected temp file to be removed after processing, still exists: %s", found)
	}
}
