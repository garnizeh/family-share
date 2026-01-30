package pipeline_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/pipeline"
	"familyshare/internal/storage"
	"familyshare/internal/testutil"
)

func TestProcessAndSave_JPEG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup test database
	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	// Setup test storage
	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	ctx := pipeline.WithSkipUploadEvent(context.Background())

	// Create test album
	album := testutil.CreateTestAlbum(t, q, "Test Album", "Description")

	// Load test JPEG image
	imgReader := testutil.LoadTestImage(t, "sample.jpg")

	// Process and save the image
	photo, err := pipeline.ProcessAndSave(ctx, db, album.ID, imgReader.(*bytes.Reader), 10<<20, storageDir)
	if err != nil {
		t.Fatalf("ProcessAndSave failed: %v", err)
	}

	// Verify photo record exists
	if photo == nil {
		t.Fatal("expected photo to be created, got nil")
	}

	if photo.AlbumID != album.ID {
		t.Errorf("expected album_id=%d, got %d", album.ID, photo.AlbumID)
	}

	if photo.Format != "webp" {
		t.Errorf("expected format=webp, got %s", photo.Format)
	}

	// Verify dimensions are within pipeline max
	if photo.Width > pipeline.MaxPipelineDimension {
		t.Errorf("width exceeds max dimension: %d > %d", photo.Width, pipeline.MaxPipelineDimension)
	}

	if photo.Height > pipeline.MaxPipelineDimension {
		t.Errorf("height exceeds max dimension: %d > %d", photo.Height, pipeline.MaxPipelineDimension)
	}

	// Verify file exists on disk
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	expectedPath := storage.PhotoPathAt(storageDir, album.ID, photo.ID, "webp", createdAt)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected file to exist at %s", expectedPath)
	}

	// Verify file is a valid WebP
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("failed to read saved file: %v", err)
	}

	// Check WebP magic bytes (RIFF...WEBP)
	if len(data) < 12 {
		t.Fatal("saved file too small to be WebP")
	}

	if string(data[0:4]) != "RIFF" {
		t.Error("file does not start with RIFF header")
	}

	if string(data[8:12]) != "WEBP" {
		t.Error("file is not WebP format")
	}
}

func TestProcessAndSave_PNG(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	ctx := pipeline.WithSkipUploadEvent(context.Background())

	album := testutil.CreateTestAlbum(t, q, "PNG Test Album", "")

	// Load test PNG image
	imgReader := testutil.LoadTestImage(t, "sample.png")

	photo, err := pipeline.ProcessAndSave(ctx, db, album.ID, imgReader.(*bytes.Reader), 10<<20, storageDir)
	if err != nil {
		t.Fatalf("ProcessAndSave failed for PNG: %v", err)
	}

	if photo.Format != "webp" {
		t.Errorf("expected PNG to be converted to webp, got %s", photo.Format)
	}

	// Verify file exists
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	expectedPath := storage.PhotoPathAt(storageDir, album.ID, photo.ID, "webp", createdAt)
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Errorf("expected WebP file to exist at %s", expectedPath)
	}
}

func TestProcessAndSave_LargeImage_Resized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	ctx := pipeline.WithSkipUploadEvent(context.Background())

	album := testutil.CreateTestAlbum(t, q, "Large Image Test", "")

	// Generate a large image (3000x2000)
	largeImg := testutil.GenerateTestImage(t, "jpeg", 3000, 2000)

	photo, err := pipeline.ProcessAndSave(ctx, db, album.ID, largeImg, 50<<20, storageDir)
	if err != nil {
		t.Fatalf("ProcessAndSave failed for large image: %v", err)
	}

	// Both dimensions should be at or below MaxPipelineDimension
	if photo.Width > pipeline.MaxPipelineDimension {
		t.Errorf("width should be resized: %d > %d", photo.Width, pipeline.MaxPipelineDimension)
	}

	if photo.Height > pipeline.MaxPipelineDimension {
		t.Errorf("height should be resized: %d > %d", photo.Height, pipeline.MaxPipelineDimension)
	}

	// Aspect ratio should be preserved (3000:2000 = 1.5:1)
	expectedRatio := 1.5
	actualRatio := float64(photo.Width) / float64(photo.Height)

	if actualRatio < expectedRatio-0.1 || actualRatio > expectedRatio+0.1 {
		t.Errorf("aspect ratio not preserved: expected ~%.2f, got %.2f", expectedRatio, actualRatio)
	}
}

func TestProcessAndSave_FileTooLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	ctx := pipeline.WithSkipUploadEvent(context.Background())

	album := testutil.CreateTestAlbum(t, q, "Size Limit Test", "")

	imgReader := testutil.LoadTestImage(t, "sample.jpg")

	// Set max bytes to 1KB (too small for any image)
	_, err := pipeline.ProcessAndSave(ctx, db, album.ID, imgReader.(*bytes.Reader), 1024, storageDir)

	// Should fail due to size limit
	if err == nil {
		t.Fatal("expected error for file too large, got nil")
	}
}

func TestProcessAndSave_Rollback_OnFileWriteFailure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	// Use invalid storage directory to force write failure
	invalidDir := "/invalid/path/that/does/not/exist"
	ctx := pipeline.WithSkipUploadEvent(context.Background())

	album := testutil.CreateTestAlbum(t, q, "Rollback Test", "")

	imgReader := testutil.LoadTestImage(t, "sample.jpg")

	// Should fail to save file
	_, err := pipeline.ProcessAndSave(ctx, db, album.ID, imgReader.(*bytes.Reader), 10<<20, invalidDir)
	if err == nil {
		t.Fatal("expected error when storage path is invalid")
	}

	// Verify no photo record was created (transaction rolled back)
	photos, err := q.ListPhotosByAlbum(ctx, sqlc.ListPhotosByAlbumParams{
		AlbumID: album.ID,
		Limit:   100,
		Offset:  0,
	})
	if err != nil {
		t.Fatalf("failed to list photos: %v", err)
	}

	if len(photos) != 0 {
		t.Errorf("expected 0 photos after rollback, got %d", len(photos))
	}
}

func TestProcessAndSave_CreatedAtTimestamp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	ctx := pipeline.WithSkipUploadEvent(context.Background())

	album := testutil.CreateTestAlbum(t, q, "Timestamp Test", "")

	before := time.Now().UTC()
	imgReader := testutil.LoadTestImage(t, "sample.jpg")

	photo, err := pipeline.ProcessAndSave(ctx, db, album.ID, imgReader.(*bytes.Reader), 10<<20, storageDir)
	if err != nil {
		t.Fatalf("ProcessAndSave failed: %v", err)
	}

	after := time.Now().UTC()

	// Verify created_at is in UTC and within expected range
	if !photo.CreatedAt.Valid {
		t.Error("created_at should be valid")
	}

	if photo.CreatedAt.Time.Before(before.Add(-1 * time.Second)) {
		t.Errorf("created_at %v is before test start %v", photo.CreatedAt.Time, before)
	}

	if photo.CreatedAt.Time.After(after.Add(1 * time.Second)) {
		t.Errorf("created_at %v is after test end %v", photo.CreatedAt.Time, after)
	}

	// Verify UTC location
	if photo.CreatedAt.Time.Location() != time.UTC {
		t.Errorf("expected created_at in UTC, got %v", photo.CreatedAt.Time.Location())
	}
}

func TestProcessAndSave_DirectoryStructure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	db, q, dbCleanup := testutil.SetupTestDB(t)
	defer dbCleanup()

	storageDir, storageCleanup := testutil.SetupTestStorage(t)
	defer storageCleanup()

	ctx := pipeline.WithSkipUploadEvent(context.Background())

	album := testutil.CreateTestAlbum(t, q, "Directory Test", "")

	imgReader := testutil.LoadTestImage(t, "sample.jpg")

	photo, err := pipeline.ProcessAndSave(ctx, db, album.ID, imgReader.(*bytes.Reader), 10<<20, storageDir)
	if err != nil {
		t.Fatalf("ProcessAndSave failed: %v", err)
	}

	// Verify directory structure follows the expected pattern
	createdAt := time.Now().UTC()
	if photo.CreatedAt.Valid {
		createdAt = photo.CreatedAt.Time.UTC()
	}
	expectedPath := storage.PhotoPathAt(storageDir, album.ID, photo.ID, "webp", createdAt)

	// Check parent directories were created
	parentDir := filepath.Dir(expectedPath)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		t.Errorf("expected parent directory to exist: %s", parentDir)
	}
}
