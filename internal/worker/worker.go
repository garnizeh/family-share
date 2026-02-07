package worker

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"familyshare/internal/config"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/pipeline"
	"familyshare/internal/storage"
)

// Worker handles background processing of uploaded photos
type Worker struct {
	db      *sql.DB
	queries *sqlc.Queries
	store   *storage.Storage
	cfg     *config.Config
	trigger chan struct{} // Channel to wake up the worker immediately
	wg      sync.WaitGroup // WaitGroup to wait for active jobs to finish
}

// NewWorker creates a new background worker
func NewWorker(db *sql.DB, store *storage.Storage, cfg *config.Config) *Worker {
	return &Worker{
		db:      db,
		queries: sqlc.New(db),
		store:   store,
		cfg:     cfg,
		trigger: make(chan struct{}, 1),
	}
}

// Start runs the background worker loop in a goroutine
func (w *Worker) Start(ctx context.Context) {
	log.Println("Worker: started background processing queue")

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		
		// Poll every 2 seconds as a fallback
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Worker: context cancelled, stopping loop")
				return
			case <-ticker.C:
				w.processBatch(ctx)
			case <-w.trigger:
				w.processBatch(ctx)
			}
		}
	}()
}

// Stop waits for the worker to finish current tasks
func (w *Worker) Stop() {
	log.Println("Worker: waiting for active jobs to finish...")
	w.wg.Wait()
	log.Println("Worker: stopped")
}

// TriggerSignal wakes up the worker to process pending jobs immediately
func (w *Worker) TriggerSignal() {
	select {
	case w.trigger <- struct{}{}:
	default:
		// already triggered
	}
}

// processBatch processes jobs until the queue is empty
func (w *Worker) processBatch(ctx context.Context) {
	for {
		// Stop if context cancelled
		if ctx.Err() != nil {
			return
		}

		// Try to pick a job
		didWork := w.processNextJob(ctx)
		if !didWork {
			return // Queue is empty, go back to sleep
		}
	}
}

func (w *Worker) processNextJob(ctx context.Context) bool {
	// 1. Get next job (atomically marks as processing)
	job, err := w.queries.GetNextPendingJob(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return false
		}
		log.Printf("Worker error checking queue: %v", err)
		return false
	}

	log.Printf("Worker: processing job %d for file %s", job.ID, job.OriginalFilename)

	// 2. Open Temp File
	f, err := os.Open(job.TempFilepath)
	if err != nil {
		w.failJob(ctx, job.ID, fmt.Sprintf("failed to open temp file: %v", err))
		return true // we did work (processed a failure), continue
	}
	
	// Ensure cleanup of filesystem
	defer func() {
		f.Close()
		os.Remove(job.TempFilepath) // We are done with it, success or fail
	}()

	fi, err := f.Stat()
	if err != nil {
		w.failJob(ctx, job.ID, "failed to stat file")
		return true
	}
	size := fi.Size()

	// 3. Process
	format := "webp"
	if w.cfg != nil && w.cfg.ImageFormat != "" {
		format = w.cfg.ImageFormat
	}

	// Use a new context for processing to ensure it doesn't get cancelled 
	// mid-way if the batch context is tight (though here we pass app ctx)
	// We inject a flag so pipeline knows context? Not strictly needed unless pipeline checks it.
	
	_, pErr := pipeline.ProcessAndSaveWithFormat(ctx, w.db, job.AlbumID, f, size, w.store.BaseDir, format)

	// 4. Update Status
	if pErr != nil {
		log.Printf("Worker: job %d failed: %v", job.ID, pErr)
		w.failJob(ctx, job.ID, pErr.Error())
	} else {
		w.completeJob(ctx, job.ID)
	}

	return true // did work
}

func (w *Worker) failJob(ctx context.Context, id int64, msg string) {
	err := w.queries.UpdateJobStatus(ctx, sqlc.UpdateJobStatusParams{
		Status:       "failed",
		ErrorMessage: sql.NullString{String: msg, Valid: true},
		ID:           id,
	})
	if err != nil {
		log.Printf("Worker: failed to update status to failed for job %d: %v", id, err)
	}
}

func (w *Worker) completeJob(ctx context.Context, id int64) {
	err := w.queries.UpdateJobStatus(ctx, sqlc.UpdateJobStatusParams{
		Status:       "completed",
		ErrorMessage: sql.NullString{},
		ID:           id,
	})
	if err != nil {
		log.Printf("Worker: failed to update status to completed for job %d: %v", id, err)
	}
}
