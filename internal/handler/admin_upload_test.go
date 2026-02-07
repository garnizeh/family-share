package handler_test

import (
	"bytes"
	"context"
	"database/sql"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/config"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/handler"
	"familyshare/internal/storage"
	"familyshare/internal/testutil"
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
	dbConn, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()
	t.Setenv("STORAGE_PATH", storageDir)

	store := storage.New(storageDir)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10}, nil)

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
		t.Fatalf("single upload status: %d, body: %s", res.StatusCode, w.Body.String())
	}
	buf := w.Body.String()
	// Update test expectation: now returns progress bar HTML, not immediate success row
	if !strings.Contains(buf, "upload-container") || !strings.Contains(buf, "Processing Photos") {
		t.Fatalf("expected progress container, got body: %s", buf)
	}

	// Wait for worker to process (since we passed nil worker, this test environment relies on Manual processing?
	// No, the tests pass nil worker, so the handler queues the job but nothing processes it.
	// We need to test that the JOB was enqueued.

	// Check queue table
	var count int
	err = dbConn.QueryRow("SELECT count(*) FROM processing_queue WHERE status = 'pending'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query queue: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 pending job, got %d", count)
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

	// Expect progress bar again
	if !strings.Contains(w2.Body.String(), "upload-container") {
		t.Fatalf("expected progress container, got: %s", w2.Body.String())
	}

	// Check queue count increases
	err = dbConn.QueryRow("SELECT count(*) FROM processing_queue WHERE status = 'pending'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	// 1 existing + 3 new = 4
	if count != 4 {
		t.Errorf("expected 4 pending jobs, got %d", count)
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
	// Even for invalid files, they are queued first, then fail in worker.
	// So response is still success (progress bar).
	err = dbConn.QueryRow("SELECT count(*) FROM processing_queue WHERE status = 'pending'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Errorf("expected 5 pending jobs, got %d", count)
	}

	// Even for invalid files, they are queued first, then fail in worker.
	// So response is still success (progress bar).
	err = dbConn.QueryRow("SELECT count(*) FROM processing_queue WHERE status = 'pending'").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 5 {
		t.Errorf("expected 5 pending jobs, got %d", count)
	}
	if w3.Result().StatusCode != 200 {
		t.Fatalf("invalid upload status: %d", w3.Result().StatusCode)
	}
	// In async mode, invalid files are queued and fail later (in worker).
	// So we don't get an immediate error message in the response.
	if !strings.Contains(w3.Body.String(), "upload-container") {
		t.Fatalf("expected progress interface, got: %s", w3.Body.String())
	}
}

func TestAdminUpload_SizeLimitRejection(t *testing.T) {
	dbConn, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	store := storage.New(storageDir)
	h := handler.New(dbConn, store, web.EmbedFS, &config.Config{RateLimitShare: 60, RateLimitAdmin: 10}, nil)

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
		_, _ = fw.Write(chunk)
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

	// In async/queue mode, files exceeding size limit are silently skipped (logged)
	// and not enqueued. The user interface just shows status.
	// We verify that no job was queued (DB check below verifies no photo created, but we should also check queue).

	var count int
	err = dbConn.QueryRow("SELECT count(*) FROM processing_queue").Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected 0 queued jobs for oversized file, got %d", count)
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
