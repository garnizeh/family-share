package storage_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"familyshare/internal/storage"
)

func TestCleanOrphanedTempFiles(t *testing.T) {
	// create temp dir
	tmp := os.TempDir()
	f1 := filepath.Join(tmp, "upload-old.tmp")
	f2 := filepath.Join(tmp, "upload-new.tmp")

	// create files
	if err := os.WriteFile(f1, []byte("old"), 0600); err != nil {
		t.Fatalf("write f1: %v", err)
	}
	if err := os.WriteFile(f2, []byte("new"), 0600); err != nil {
		t.Fatalf("write f2: %v", err)
	}

	// set modification times: f1 old, f2 now
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(f1, old, old); err != nil {
		t.Fatalf("chtimes f1: %v", err)
	}

	// run janitor with cutoff 15 minutes -> should remove f1 only
	if err := storage.CleanOrphanedTempFiles(15 * time.Minute); err != nil {
		t.Fatalf("clean: %v", err)
	}

	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Fatalf("expected f1 removed, stat err: %v", err)
	}
	if _, err := os.Stat(f2); err != nil {
		t.Fatalf("expected f2 to remain: %v", err)
	}

	// cleanup
	_ = os.Remove(f2)
}
