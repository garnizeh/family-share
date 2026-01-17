package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/pipeline"
)

// UploadResult is passed to the HTMX partial for each uploaded file.
type UploadResult struct {
	Filename string
	PhotoID  int64
	Error    error
}

// AdminUploadPhotos handles multipart photo uploads for an album.
func (h *Handler) AdminUploadPhotos(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	albumIDStr := chi.URLParam(r, "id")
	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	// Total upload limit (100MB)
	const maxTotal = int64(100 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, maxTotal)
	if err := r.ParseMultipartForm(maxTotal); err != nil {
		http.Error(w, "failed to parse multipart form", http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["photos"]
	if len(files) == 0 {
		http.Error(w, "no files uploaded", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Process each file and emit an HTMX partial per file. Flush after each
	// so the client can show incremental progress.
	flusher, _ := w.(http.Flusher)

	for _, fh := range files {
		var result UploadResult
		result.Filename = fh.Filename

		// Per-file size guard
		const maxPerFile = int64(25 << 20) // 25MB
		if fh.Size > 0 && fh.Size > maxPerFile {
			result.Error = fmt.Errorf("file too large")
			h.RenderTemplate(w, "upload_row.html", result)
			if flusher != nil {
				flusher.Flush()
			}
			continue
		}

		file, err := fh.Open()
		if err != nil {
			result.Error = fmt.Errorf("open failed: %w", err)
			h.RenderTemplate(w, "upload_row.html", result)
			if flusher != nil {
				flusher.Flush()
			}
			continue
		}

		// Read into memory then create a ReadSeeker for the pipeline.
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, file); err != nil {
			result.Error = fmt.Errorf("read failed: %w", err)
			file.Close()
			h.RenderTemplate(w, "upload_row.html", result)
			if flusher != nil {
				flusher.Flush()
			}
			continue
		}
		file.Close()

		reader := bytes.NewReader(buf.Bytes())

		photo, err := pipeline.ProcessAndSave(context.WithValue(ctx, "admin-upload", true), h.db, albumID, reader, int64(buf.Len()))
		if err != nil {
			result.Error = err
		} else {
			result.PhotoID = photo.ID
		}

		// Render HTMX partial for this file
		if err := h.RenderTemplate(w, "upload_row.html", result); err != nil {
			// If template fails, write a fallback
			http.Error(w, "template render error", http.StatusInternalServerError)
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}
}
