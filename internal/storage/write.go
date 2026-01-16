package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// EnsureDir creates directory structure with proper permissions.
func EnsureDir(path string) error {
	return os.MkdirAll(path, 0o755)
}

// AtomicWrite writes data to path atomically using a temp file in the same directory.
func AtomicWrite(path string, data io.Reader) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("ensure dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	// ensure cleanup of tmp on error
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	if _, err := io.Copy(tmp, data); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("rename temp to final: %w", err)
	}

	return nil
}
