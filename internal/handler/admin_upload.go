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

	"github.com/go-chi/chi/v5"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/pipeline"
)

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

// AdminUploadStatus returns htmx partial with progress of background processing
func (h *Handler) AdminUploadStatus(w http.ResponseWriter, r *http.Request) {
	albumIDStr := r.URL.Query().Get("album_id")
	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil {
		// If called from URL param
		albumIDStr = chi.URLParam(r, "id")
		albumID, err = strconv.ParseInt(albumIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid album ID", http.StatusBadRequest)
			return
		}
	}

	status, err := h.queries.GetQueueStatus(r.Context(), albumID)
	if err != nil {
		log.Printf("failed to get queue status: %v", err)
		http.Error(w, "Check failed", http.StatusInternalServerError)
		return
	}

	// Calculate statistics
	total := 0.0
	if status.PendingCount.Valid {
		total += status.PendingCount.Float64
	}
	if status.ProcessingCount.Valid {
		total += status.ProcessingCount.Float64
	}
	if status.CompletedCount.Valid {
		total += status.CompletedCount.Float64
	}
	if status.FailedCount.Valid {
		total += status.FailedCount.Float64
	}

	processed := 0.0
	if status.CompletedCount.Valid {
		processed += status.CompletedCount.Float64
	}
	if status.FailedCount.Valid {
		processed += status.FailedCount.Float64
	}

	percent := 0
	if total > 0 {
		percent = int((processed / total) * 100)
	}

	data := struct {
		Album struct {
			ID int64
		}
		Stats struct {
			PendingCount    int64
			ProcessingCount int64
			CompletedCount  int64
			FailedCount     int64
			Percent         int
		}
	}{
		Album: struct{ ID int64 }{ID: albumID},
		Stats: struct {
			PendingCount    int64
			ProcessingCount int64
			CompletedCount  int64
			FailedCount     int64
			Percent         int
		}{
			PendingCount:    int64(status.PendingCount.Float64),
			ProcessingCount: int64(status.ProcessingCount.Float64),
			CompletedCount:  int64(status.CompletedCount.Float64),
			FailedCount:     int64(status.FailedCount.Float64),
			Percent:         percent,
		},
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.RenderTemplate(w, "upload_progress", data); err != nil {
		log.Printf("failed to render upload_progress: %v", err)
	}
}

// AdminUploadPhotos handles multipart photo uploads for an album
// It now queues files for background processing instead of processing them synchronously
func (h *Handler) AdminUploadPhotos(w http.ResponseWriter, r *http.Request) {
	albumIDStr := chi.URLParam(r, "id")
	albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid album id", http.StatusBadRequest)
		return
	}

	// Total upload limit increased for batching (500MB)
	const maxTotal = int64(500 << 20)
	r.Body = http.MaxBytesReader(w, r.Body, maxTotal)

	mr, err := r.MultipartReader()
	if err != nil {
		http.Error(w, "failed to read multipart", http.StatusBadRequest)
		return
	}

	// Determine temp dir
	tmpBaseDir := os.Getenv("TEMP_UPLOAD_DIR")
	if tmpBaseDir == "" {
		tmpBaseDir = os.TempDir()
	}
	_ = os.MkdirAll(tmpBaseDir, 0700)

	const maxPerFile = int64(25 << 20) // 25MB per file

	filesQueued := 0

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("multipart read error: %v", err)
			continue
		}

		if part.FormName() != "photos" {
			part.Close()
			continue
		}

		filename := part.FileName()
		if filename == "" {
			part.Close()
			continue
		}

		// Create temp file
		tmp, err := os.CreateTemp(tmpBaseDir, "upload-*.tmp")
		if err != nil {
			log.Printf("failed to create temp file for %s: %v", filename, err)
			part.Close()
			continue
		}

		// Copy file content
		n, err := io.Copy(tmp, io.LimitReader(part, maxPerFile+1))
		part.Close()
		tmp.Close() // Close immediately after writing

		if err != nil {
			log.Printf("failed to save temp file for %s: %v", filename, err)
			os.Remove(tmp.Name())
			continue
		}

		if n > maxPerFile {
			log.Printf("file too large: %s", filename)
			os.Remove(tmp.Name())
			continue
		}

		// Enqueue the job
		_, err = h.queries.EnqueueJob(context.Background(), sqlc.EnqueueJobParams{
			AlbumID:          albumID,
			OriginalFilename: filename,
			TempFilepath:     tmp.Name(),
		})
		if err != nil {
			log.Printf("failed to enqueue job for %s: %v", filename, err)
			os.Remove(tmp.Name())
			continue
		}

		filesQueued++
	}

	// Trigger worker to start processing immediately
	if h.worker != nil && filesQueued > 0 {
		h.worker.TriggerSignal()
	}

	// Render the progress bar immediately
	// We pass the request to AdminUploadStatus to reuse logic
	h.AdminUploadStatus(w, r)
}
