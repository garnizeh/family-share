package storage

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Cleanup is a simple helper to track temporary paths and remove them.
type Cleanup struct {
	paths []string
}

// Add registers a path for later cleanup.
func (c *Cleanup) Add(path string) {
	c.paths = append(c.paths, path)
}

// Execute removes all registered paths. It is safe to call multiple times.
// Returns the first non-ignorable error encountered, or nil.
func (c *Cleanup) Execute() error {
	var firstErr error
	for _, p := range c.paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	c.paths = nil
	return firstErr
}

// CleanOrphanedTempFiles removes temp upload files older than maxAge from the system temp dir.
// It only touches files matching the prefix/suffix pattern used by uploads: "upload-*.tmp".
func CleanOrphanedTempFiles(maxAge time.Duration) error {
	tmpDir := os.TempDir()
	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return err
	}

	cutoff := time.Now().UTC().Add(-maxAge)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "upload-") || !strings.HasSuffix(name, ".tmp") {
			continue
		}
		full := filepath.Join(tmpDir, name)
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(full)
		}
	}
	return nil
}
