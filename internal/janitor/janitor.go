package janitor

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/storage"
)

// Janitor handles periodic cleanup of expired data and orphaned files
type Janitor struct {
	db          *sql.DB
	queries     *sqlc.Queries
	storagePath string
	interval    time.Duration
	stopChan    chan struct{}
	doneChan    chan struct{}
}

// Config holds janitor configuration
type Config struct {
	DB          *sql.DB
	StoragePath string
	Interval    time.Duration
}

// New creates a new Janitor instance
func New(cfg Config) *Janitor {
	if cfg.Interval == 0 {
		cfg.Interval = 6 * time.Hour // default to 6 hours
	}

	return &Janitor{
		db:          cfg.DB,
		queries:     sqlc.New(cfg.DB),
		storagePath: cfg.StoragePath,
		interval:    cfg.Interval,
		stopChan:    make(chan struct{}),
		doneChan:    make(chan struct{}),
	}
}

// Start begins the cleanup scheduler in a goroutine
func (j *Janitor) Start(ctx context.Context) {
	go j.run(ctx)
}

// Stop gracefully stops the janitor
func (j *Janitor) Stop() {
	close(j.stopChan)
	<-j.doneChan // wait for cleanup to finish
}

// run is the main loop that runs cleanup tasks
func (j *Janitor) run(ctx context.Context) {
	defer close(j.doneChan)

	// Run cleanup immediately on startup
	j.runCleanup(ctx)

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			j.runCleanup(ctx)
		case <-j.stopChan:
			log.Println("Janitor: received stop signal, shutting down...")
			return
		case <-ctx.Done():
			log.Println("Janitor: context cancelled, shutting down...")
			return
		}
	}
}

// runCleanup executes all cleanup tasks
func (j *Janitor) runCleanup(ctx context.Context) {
	log.Println("Janitor: starting cleanup cycle...")
	start := time.Now().UTC()

	j.deleteExpiredSessions(ctx)
	j.deleteExpiredShareLinks(ctx)
	j.deleteOrphanedPhotos(ctx)
	j.deleteOldActivityEvents(ctx)
	j.cleanupTempFiles()

	duration := time.Since(start)
	log.Printf("Janitor: cleanup cycle completed in %v", duration)
}

// deleteExpiredSessions removes expired sessions from the database
func (j *Janitor) deleteExpiredSessions(ctx context.Context) {
	err := j.queries.DeleteExpiredSessions(ctx)
	if err != nil {
		log.Printf("Janitor: failed to delete expired sessions: %v", err)
		return
	}
	log.Println("Janitor: deleted expired sessions")
}

// deleteExpiredShareLinks removes expired and revoked share links
func (j *Janitor) deleteExpiredShareLinks(ctx context.Context) {
	links, err := j.queries.DeleteExpiredShareLinks(ctx)
	if err != nil {
		log.Printf("Janitor: failed to delete expired share links: %v", err)
		return
	}
	
	if len(links) > 0 {
		log.Printf("Janitor: deleted %d expired/revoked share links", len(links))
	}
}

// deleteOrphanedPhotos removes photos whose albums no longer exist and deletes their files
func (j *Janitor) deleteOrphanedPhotos(ctx context.Context) {
	photos, err := j.queries.DeleteOrphanedPhotos(ctx)
	if err != nil {
		log.Printf("Janitor: failed to delete orphaned photos: %v", err)
		return
	}

	if len(photos) == 0 {
		return
	}

	log.Printf("Janitor: found %d orphaned photos, deleting files...", len(photos))

	deletedCount := 0
	for _, photo := range photos {
		// Delete main photo file
		photoPath := storage.PhotoPath(j.storagePath, photo.AlbumID, photo.ID, photo.Format)
		if err := j.deleteFile(photoPath); err != nil {
			log.Printf("Janitor: failed to delete photo file %s: %v", photoPath, err)
		} else {
			deletedCount++
		}

		// Delete thumbnail if exists
		thumbPath := storage.ThumbnailPath(j.storagePath, photo.AlbumID, photo.ID)
		if err := j.deleteFile(thumbPath); err != nil {
			// Thumbnails are optional, don't log error if not found
			if !os.IsNotExist(err) {
				log.Printf("Janitor: failed to delete thumbnail %s: %v", thumbPath, err)
			}
		}
	}

	log.Printf("Janitor: deleted %d orphaned photo files", deletedCount)
	
	// Clean up empty directories
	j.cleanupEmptyDirs()
}

// deleteFile removes a file from disk
func (j *Janitor) deleteFile(path string) error {
	return os.Remove(path)
}

// cleanupEmptyDirs removes empty directories in the photos directory structure
func (j *Janitor) cleanupEmptyDirs() {
	photosDir := filepath.Join(j.storagePath, "photos")
	
	// Walk the directory tree from bottom to top
	filepath.Walk(photosDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip on error
		}
		
		// Skip if not a directory or if it's the root photos directory
		if !info.IsDir() || path == photosDir {
			return nil
		}
		
		// Try to remove empty directories (will fail if not empty)
		if err := os.Remove(path); err == nil {
			log.Printf("Janitor: removed empty directory: %s", path)
		}
		
		return nil
	})
}

// cleanupTempFiles removes orphaned temporary upload files older than 15 minutes
func (j *Janitor) cleanupTempFiles() {
	if err := storage.CleanOrphanedTempFiles(15 * time.Minute); err != nil {
		log.Printf("Janitor: failed to cleanup temp files: %v", err)
	}
}

// deleteOldActivityEvents removes activity events older than 90 days
func (j *Janitor) deleteOldActivityEvents(ctx context.Context) {
	ninetyDaysAgo := time.Now().UTC().Add(-90 * 24 * time.Hour)
	
	err := j.queries.DeleteOldActivityEvents(ctx, sql.NullTime{Time: ninetyDaysAgo, Valid: true})
	if err != nil {
		log.Printf("Janitor: failed to delete old activity events: %v", err)
		return
	}
	log.Println("Janitor: deleted old activity events (90+ days)")
}
