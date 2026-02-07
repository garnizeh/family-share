package worker

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"familyshare/internal/config"
	"familyshare/internal/db/sqlc"
	"familyshare/internal/storage"
	"familyshare/internal/testutil"

	_ "modernc.org/sqlite"
)

func TestWorker_Integration(t *testing.T) {
	// 1. Setup DB and Storage
	db, queries, cleanupDB := testutil.SetupTestDB(t)
	defer cleanupDB()

	// Use specific temporary directory for this test
	tempDir := t.TempDir()
	store := storage.New(tempDir)

	cfg := &config.Config{
		ImageFormat: "webp",
	}

	// 2. Create Album
	album, err := queries.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{
		Title: "Test Worker Album",
	})
	if err != nil {
		t.Fatalf("failed to create album: %v", err)
	}

	// 3. Create Dummy Image File
	tmpFile := filepath.Join(tempDir, "test_image.jpg")
	// Create a valid dummy image (using a small valid JPEG header or using testutil helper if available)
	// For simplicity, we create a text file expecting it to fail the pipeline (NotAnImage),
	// verifying the worker handles failure correctly.
	// If we want success, we need a real image bytes.
	if err := os.WriteFile(tmpFile, []byte("not an image"), 0644); err != nil {
		t.Fatalf("failed to create dummy file: %v", err)
	}

	// 4. Enqueue Job
	job, err := queries.EnqueueJob(context.Background(), sqlc.EnqueueJobParams{
		AlbumID:          album.ID,
		OriginalFilename: "test_image.jpg",
		TempFilepath:     tmpFile,
	})
	if err != nil {
		t.Fatalf("failed to enqueue job: %v", err)
	}

	// 5. Initialize Worker
	w := NewWorker(db, store, cfg)

	// 6. Start Worker
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	// 7. Trigger Logic (Wait for processing)
	// Monitor DB until status changes
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	var updatedJob sqlc.ProcessingQueue
	done := false

	for !done {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for worker to process job")
		case <-ticker.C:
			// check status
			err := db.QueryRowContext(context.Background(), "SELECT id, status, error_message FROM processing_queue WHERE id = ?", job.ID).Scan(&updatedJob.ID, &updatedJob.Status, &updatedJob.ErrorMessage)
			if err != nil {
				t.Fatalf("failed to query job: %v", err)
			}

			if updatedJob.Status == "completed" || updatedJob.Status == "failed" {
				done = true
			}
		}
	}

	// 8. Cleanup Worker
	cancel()
	w.Stop()

	// 9. Assertions
	// Since we uploaded "not an image", we expect "failed" status
	if updatedJob.Status != "failed" {
		t.Errorf("expected status 'failed', got '%s'", updatedJob.Status)
	}
	if !updatedJob.ErrorMessage.Valid || updatedJob.ErrorMessage.String == "" {
		t.Error("expected error message, got empty")
	}

	// Verify temp file is gone (worker should cleanup)
	if _, err := os.Stat(tmpFile); !os.IsNotExist(err) {
		t.Error("temp file should be deleted after processing")
	}
}

func TestWorker_TriggerSignal(t *testing.T) {
	// Setup DB and Storage
	db, _, cleanupDB := testutil.SetupTestDB(t)
	defer cleanupDB()

	tempDir := t.TempDir()
	store := storage.New(tempDir)
	cfg := &config.Config{ImageFormat: "webp"}
	w := NewWorker(db, store, cfg)

	// We want to test that TriggerSignal wakes up the worker WITHOUT waiting for the ticker.
	// But in integration tests, racing time is unreliable.
	// We can check if `trigger` channel has value.

	// Ensure channel is empty initially
	select {
	case <-w.trigger:
		t.Fatal("expected trigger channel to be empty")
	default:
	}

	w.TriggerSignal()

	// Verify channel received signal
	select {
	case <-w.trigger:
		// Success
	default:
		t.Fatal("expected trigger channel to have signal")
	}

	// Verify non-blocking: calling it again when full shouldn't block
	w.TriggerSignal() // Fill it (size 1 was consumed? No, NewWorker makes size 1)
	// Wait, we consumed it above. So it's empty. Now size 1.
	w.TriggerSignal() // Should be full now.

	done := make(chan bool)
	go func() {
		w.TriggerSignal() // This should NOT block even if full, due to select default
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("TriggerSignal blocked when channel is full")
	}
}

func TestWorker_completeJob(t *testing.T) {
	// Setup DB
	db, queries, cleanupDB := testutil.SetupTestDB(t)
	defer cleanupDB()

	// Create setup data: Album and Job
	album, err := queries.CreateAlbum(context.Background(), sqlc.CreateAlbumParams{Title: "Test"})
	if err != nil {
		t.Fatalf("create album: %v", err)
	}

	job, err := queries.EnqueueJob(context.Background(), sqlc.EnqueueJobParams{
		AlbumID:          album.ID,
		OriginalFilename: "foo.jpg",
		TempFilepath:     "/tmp/foo",
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	// Manually set status to 'processing' to simulate active work
	// Note: depending on your schema/logic, Enqueue sets it to 'pending'.
	// completeJob should work regardless, but let's be realistic.

	tempDir := t.TempDir()
	store := storage.New(tempDir)
	cfg := &config.Config{}
	w := NewWorker(db, store, cfg)

	// Action
	w.completeJob(context.Background(), job.ID)

	// Verify
	var status string
	var errMsg sql.NullString
	err = db.QueryRow("SELECT status, error_message FROM processing_queue WHERE id = ?", job.ID).Scan(&status, &errMsg)
	if err != nil {
		t.Fatalf("query job: %v", err)
	}

	if status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", status)
	}
	if errMsg.Valid {
		t.Errorf("expected no error message, got '%s'", errMsg.String)
	}
}
