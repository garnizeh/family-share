package handler_test

import (
"familyshare/internal/config"
	"bytes"
	"context"
	"database/sql"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/web"
)

// helper to create a small jpeg image in memory
func makeJPEG(t *testing.T, w, h int) *bytes.Buffer {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// fill with a solid color
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 0x80, G: 0x90, B: 0xA0, A: 0xFF})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatalf("encode jpeg: %v", err)
	}
	return &buf
}

// attachFile adds a file part to the multipart writer with given fieldname and contents
func attachFile(t *testing.T, mw *multipart.Writer, field, filename string, r io.Reader) {
	t.Helper()
	fw, err := mw.CreateFormFile(field, filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(fw, r); err != nil {
		t.Fatalf("copy file: %v", err)
	}
}

func TestAdminUpload_SingleAndBatchAndInvalid(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	// storage path in a temp dir
	tmpd, err := os.MkdirTemp("", "fs-test-storage-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(tmpd)
	os.Setenv("STORAGE_PATH", tmpd)

	store := storage.New(tmpd)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	q := sqlc.New(dbConn)
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{Title: "test", Description: sql.NullString{String: "", Valid: false}})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	// --- Single upload ---
	singleBuf := makeJPEG(t, 32, 32)
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	attachFile(t, mw, "photos", "one.jpg", singleBuf)
	mw.Close()

	req := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	// set chi URL param
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()
	h.AdminUploadPhotos(w, req)
	res := w.Result()
	if res.StatusCode != 200 {
		t.Fatalf("single upload status: %d", res.StatusCode)
	}
	buf := w.Body.String()
	if !strings.Contains(buf, "Successfully uploaded (ID:") {
		t.Fatalf("expected uploaded partial, got body: %s", buf)
	}

	// verify a photo file was written under STORAGE_PATH
	// photos are stored under storage.PhotoPath(base, albumID, photoID, ext)
	// find any file under tmpd recursively
	var found bool
	filepath.Walk(tmpd, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			found = true
		}
		return nil
	})
	if !found {
		t.Fatalf("expected stored photo file in %s", tmpd)
	}

	// --- Batch upload (3 files) ---
	var body2 bytes.Buffer
	mw2 := multipart.NewWriter(&body2)
	for i := 0; i < 3; i++ {
		attachFile(t, mw2, "photos", "b"+strconv.Itoa(i)+".jpg", makeJPEG(t, 24, 24))
	}
	mw2.Close()

	req2 := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", &body2)
	req2.Header.Set("Content-Type", mw2.FormDataContentType())
	rc2 := chi.NewRouteContext()
	rc2.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req2 = req2.WithContext(context.WithValue(req2.Context(), chi.RouteCtxKey, rc2))

	w2 := httptest.NewRecorder()
	h.AdminUploadPhotos(w2, req2)
	if w2.Result().StatusCode != 200 {
		t.Fatalf("batch upload status: %d", w2.Result().StatusCode)
	}
	if strings.Count(w2.Body.String(), "Successfully uploaded (ID:") < 3 {
		t.Fatalf("expected 3 uploaded partials, got: %s", w2.Body.String())
	}

	// --- Invalid file upload ---
	var body3 bytes.Buffer
	mw3 := multipart.NewWriter(&body3)
	// text file
	attachFile(t, mw3, "photos", "not-image.txt", bytes.NewBufferString("hello world"))
	mw3.Close()

	req3 := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", &body3)
	req3.Header.Set("Content-Type", mw3.FormDataContentType())
	rc3 := chi.NewRouteContext()
	rc3.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req3 = req3.WithContext(context.WithValue(req3.Context(), chi.RouteCtxKey, rc3))

	w3 := httptest.NewRecorder()
	h.AdminUploadPhotos(w3, req3)
	if w3.Result().StatusCode != 200 {
		t.Fatalf("invalid upload status: %d", w3.Result().StatusCode)
	}
	if !strings.Contains(w3.Body.String(), "Error:") {
		t.Fatalf("expected failure partial, got: %s", w3.Body.String())
	}
}

func TestAdminUpload_SizeLimitRejection(t *testing.T) {
	dbConn, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("InitDB: %v", err)
	}
	defer dbConn.Close()

	tmpd, err := os.MkdirTemp("", "fs-test-storage-")
	if err != nil {
		t.Fatalf("mktemp: %v", err)
	}
	defer os.RemoveAll(tmpd)
	os.Setenv("STORAGE_PATH", tmpd)

	store := storage.New(tmpd)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10})

	q := sqlc.New(dbConn)
	album, err := q.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{Title: "test", Description: sql.NullString{String: "", Valid: false}})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	// Create a file that exceeds per-file limit (25MB)
	// We'll create a fake large payload to simulate size check
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)

	// Create form file header
	fw, err := mw.CreateFormFile("photos", "toolarge.jpg")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	// Write 26MB of data (exceeds 25MB limit)
	// Use a repeating pattern to save memory during test
	chunk := bytes.Repeat([]byte("x"), 1024*1024) // 1MB chunk
	for range 26 {
		fw.Write(chunk)
	}
	mw.Close()

	req := httptest.NewRequest("POST", "/admin/albums/"+strconv.FormatInt(album.ID, 10)+"/photos", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rc := chi.NewRouteContext()
	rc.URLParams.Add("id", strconv.FormatInt(album.ID, 10))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rc))

	w := httptest.NewRecorder()
	h.AdminUploadPhotos(w, req)

	res := w.Result()
	if res.StatusCode != 200 {
		t.Fatalf("expected 200 with error partial, got: %d", res.StatusCode)
	}

	responseBody := w.Body.String()
	if !strings.Contains(responseBody, "file too large") {
		t.Fatalf("expected 'file too large' error, got: %s", responseBody)
	}

	// Verify no photo was created in DB
	photos, err := q.ListPhotosByAlbum(context.Background(), sqlc.ListPhotosByAlbumParams{
		AlbumID: album.ID,
		Limit:   10,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("list photos: %v", err)
	}
	if len(photos) > 0 {
		t.Fatalf("expected no photos created for oversized file, got %d", len(photos))
	}
}
