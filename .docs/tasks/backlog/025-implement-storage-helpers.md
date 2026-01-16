# Task 025: Implement Storage Path and File Helpers

**Milestone:** Storage & Pipeline  
**Points:** 1 (5 hours)  
**Dependencies:** 010  
**Branch:** `feat/storage-helpers`  
**Labels:** `storage`, `infrastructure`

## Description
Create utility functions for storage path generation, atomic file writes, and safe cleanup. These will be used by the image pipeline.

## Acceptance Criteria
- [ ] Path generation follows TDD layout: `data/photos/{yyyy}/{mm}/{album_id}/{photo_id}.{ext}`
- [ ] Atomic write implemented (write to temp, rename on success)
- [ ] Safe cleanup for temporary files (defer-based)
- [ ] Directory creation with proper permissions
- [ ] File existence checks and error handling

## Files to Add/Modify
- `internal/storage/paths.go` — path generation logic
- `internal/storage/write.go` — atomic write operations
- `internal/storage/cleanup.go` — temp file cleanup
- `internal/storage/storage_test.go` — unit tests

## Key Functions
```go
// PhotoPath returns the storage path for a photo
func PhotoPath(baseDir string, albumID, photoID int64, format string) string

// ThumbnailPath returns the storage path for a thumbnail
func ThumbnailPath(baseDir string, albumID, photoID int64) string

// AtomicWrite writes data to path atomically (temp + rename)
func AtomicWrite(path string, data io.Reader) error

// EnsureDir creates directory structure with permissions
func EnsureDir(path string) error

// Cleanup removes temporary files, safe to call multiple times
type Cleanup struct {
    paths []string
}
func (c *Cleanup) Add(path string)
func (c *Cleanup) Execute() error
```

## Tests Required
- [ ] Unit test: PhotoPath generates correct structure
- [ ] Unit test: AtomicWrite succeeds and creates file
- [ ] Unit test: AtomicWrite fails halfway, no partial file left
- [ ] Unit test: Cleanup removes all registered temp files
- [ ] Unit test: EnsureDir creates nested directories

## PR Checklist
- [ ] All functions have doc comments
- [ ] Error messages are descriptive
- [ ] Tests pass: `go test ./internal/storage/... -v`
- [ ] Edge cases handled (empty albumID, invalid paths, permission errors)
- [ ] No hardcoded paths (use configurable base directory)

## Git Workflow
```bash
git checkout -b feat/storage-helpers
# Implement helpers
go test ./internal/storage/... -v -cover
git add internal/storage/
git commit -m "feat: implement storage path and atomic write helpers"
git push origin feat/storage-helpers
# Open PR: "Add storage helpers for path generation and atomic writes"
```

## Notes
- Use `os.MkdirAll` for directory creation
- Atomic write pattern: `ioutil.TempFile` → write → `os.Rename`
- Paths should be OS-agnostic (`filepath.Join`)
- Consider using year/month subdirectories to avoid too many files in one dir
