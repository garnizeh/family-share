package storage

import (
	"os"
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
