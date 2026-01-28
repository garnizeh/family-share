package testutil

import (
	"os"
	"testing"
)

// SetupTestStorage creates a temporary directory for file storage during tests.
// Returns the directory path and a cleanup function that should be deferred.
func SetupTestStorage(t *testing.T) (string, func()) {
	t.Helper()

	dir, err := os.MkdirTemp("", "familyshare-test-*")
	if err != nil {
		t.Fatalf("failed to create test storage directory: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return dir, cleanup
}