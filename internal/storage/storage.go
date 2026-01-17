package storage

// Storage is a minimal storage wrapper for the base directory used by helpers.
// Keep it lightweight so handlers can depend on it where appropriate.
type Storage struct {
	BaseDir string
}

// New creates a new Storage instance with the provided base directory.
func New(baseDir string) *Storage {
	return &Storage{BaseDir: baseDir}
}
