package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"familyshare/internal/pipeline"
)

// UploadResult is passed to the HTMX partial for each uploaded file.
type UploadResult struct {
	Filename string
	PhotoID  int64
	Error    string
}

var errUploadTooLarge = errors.New("upload exceeds per-file limit")

func friendlyUploadError(err error, maxPerFile int64) string {
	if err == nil {
		return ""
	}

	switch {
	case errors.Is(err, errUploadTooLarge), errors.Is(err, pipeline.ErrTooLarge):
		return fmt.Sprintf("File is too large. Max %dMB.", maxPerFile>>20)
	case errors.Is(err, pipeline.ErrNotAnImage):
		return "Unsupported file type. Please upload a JPG, PNG, WebP, GIF, or AVIF."
	case errors.Is(err, pipeline.ErrInvalidDimensions):
		return fmt.Sprintf("Image dimensions are invalid. Max %dx%d pixels.", pipeline.MaxDimension, pipeline.MaxDimension)
	case errors.Is(err, pipeline.ErrDecodeFailed):
		return "We couldn't read that image. It may be corrupted."
	default:
		return "Upload failed. Please try again."
	}
}

// AdminUploadPhotos handles multipart photo uploads for an album.
func (h *Handler) AdminUploadPhotos(w http.ResponseWriter, r *http.Request) {
	// Set per-request timeout to avoid hung uploads
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
	defer cancel()

	albumIDStr := chi.URLParam(r, "id")
	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	// Total upload limit (100MB)
	const maxTotal = int64(100 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, maxTotal)

	// Use MultipartReader to stream parts to disk
	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "failed to read multipart", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	flusher, _ := w.(http.Flusher)

	// Determine temp dir: prefer TEMP_UPLOAD_DIR env var (for tests/custom paths), fallback to system temp dir
	tmpBaseDir := os.Getenv("TEMP_UPLOAD_DIR")
	if tmpBaseDir == "" {
		tmpBaseDir = os.TempDir()
	}
	// Ensure directory exists
	_ = os.MkdirAll(tmpBaseDir, 0700)

	const maxPerFile = int64(25 << 20) // 25MB

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			// log and continue to next part
			continue
		}

		if part.FormName() != "photos" {
			part.Close()
			continue
		}

		var result UploadResult
		result.Filename = part.FileName()

		// Create temp file
		tmp, err := os.CreateTemp(tmpBaseDir, "upload-*.tmp")
		if err != nil {
			log.Printf("upload temp file creation failed for %s: %v", result.Filename, err)
			result.Error = friendlyUploadError(err, maxPerFile)
			h.RenderTemplate(w, "upload_row.html", result)
			if flusher != nil {
				flusher.Flush()
			}
			part.Close()
			continue
		}
		// Restrictive permissions
		_ = tmp.Chmod(0600)

		// Copy with size limit
		n, err := io.Copy(tmp, io.LimitReader(part, maxPerFile+1))
		if err != nil {
			copyErr := fmt.Errorf("read failed: %w", err)
			log.Printf("upload read failed for %s: %v", result.Filename, copyErr)
			tmp.Close()
			os.Remove(tmp.Name())
			result.Error = friendlyUploadError(copyErr, maxPerFile)
			h.RenderTemplate(w, "upload_row.html", result)
			if flusher != nil {
				flusher.Flush()
			}
			part.Close()
			continue
		}

		if n > maxPerFile {
			log.Printf("upload too large for %s: %d bytes", result.Filename, n)
			tmp.Close()
			os.Remove(tmp.Name())
			result.Error = friendlyUploadError(errUploadTooLarge, maxPerFile)
			h.RenderTemplate(w, "upload_row.html", result)
			if flusher != nil {
				flusher.Flush()
			}
			part.Close()
			continue
		}

		// Seek to beginning for pipeline
		if _, err := tmp.Seek(0, 0); err != nil {
			seekErr := fmt.Errorf("seek failed: %w", err)
			log.Printf("upload seek failed for %s: %v", result.Filename, seekErr)
			tmp.Close()
			os.Remove(tmp.Name())
			result.Error = friendlyUploadError(seekErr, maxPerFile)
			h.RenderTemplate(w, "upload_row.html", result)
			if flusher != nil {
				flusher.Flush()
			}
			part.Close()
			continue
		}

		// Process through pipeline
		format := "webp"
		if h.config != nil && h.config.ImageFormat != "" {
			format = h.config.ImageFormat
		}
		photo, err := pipeline.ProcessAndSaveWithFormat(context.WithValue(ctx, "admin-upload", true), h.db, albumID, tmp, n, h.storage.BaseDir, format)

		// Cleanup temp file always
		tmp.Close()
		os.Remove(tmp.Name())
		part.Close()

		if err != nil {
			log.Printf("upload pipeline failed for %s: %v", result.Filename, err)
			result.Error = friendlyUploadError(err, maxPerFile)
		} else {
			result.PhotoID = photo.ID
		}

		// Render HTMX partial for this file
		if err := h.RenderTemplate(w, "upload_row.html", result); err != nil {
			http.Error(w, "template render error", http.StatusInternalServerError)
			return
		}
		if flusher != nil {
			flusher.Flush()
		}
	}
}
