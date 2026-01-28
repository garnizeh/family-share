# Task 120: Core Integration Tests and Fixtures

**Milestone:** Tests & CI  
**Points:** 2 (8 hours)  
**Dependencies:** 050, 080, 090  
**Branch:** `feat/integration-tests`  
**Labels:** `testing`, `quality`

## Description
Write integration tests for critical workflows using temporary SQLite databases and test fixtures. Cover upload pipeline, share link access, and admin auth.

## Acceptance Criteria
- [ ] Test fixtures: sample images (JPEG, PNG), test data SQL
- [ ] Integration test: full upload pipeline (multipart → save → DB)
- [ ] Integration test: share link creation and view tracking
- [ ] Integration test: admin login flow
- [ ] Integration test: album CRUD operations
- [ ] Tests use temporary DB and storage (cleanup after)

## Files to Add/Modify
- `internal/testutil/fixtures.go` — test helpers and fixtures
- `internal/testutil/db.go` — temp DB setup/teardown
- `internal/handler/handler_integration_test.go` — integration tests
- `internal/pipeline/pipeline_integration_test.go` — pipeline tests
- `testdata/images/` — sample JPEG, PNG files

## Test Helpers
```go
// SetupTestDB creates a temporary SQLite DB with migrations
func SetupTestDB(t *testing.T) (*sql.DB, func()) {
    db, _ := sql.Open("sqlite", ":memory:")
    runMigrations(db)
    return db, func() { db.Close() }
}

// SetupTestStorage creates a temp directory
func SetupTestStorage(t *testing.T) (string, func()) {
    dir, _ := os.MkdirTemp("", "familyshare-test-")
    return dir, func() { os.RemoveAll(dir) }
}

// LoadTestImage returns a test JPEG as io.Reader
func LoadTestImage(t *testing.T, name string) io.Reader
```

## Sample Integration Test
```go
func TestUploadPipeline(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()
    
    storagePath, storageCleanup := SetupTestStorage(t)
    defer storageCleanup()
    
    // Create album
    album := createTestAlbum(db, "Test Album")
    
    // Upload photo
    imageData := LoadTestImage(t, "sample.jpg")
    photo, err := pipeline.ProcessAndSave(ctx, db, album.ID, imageData, 10<<20)
    
    assert.NoError(t, err)
    assert.NotNil(t, photo)
    
    // Verify file exists
    assert.FileExists(t, filepath.Join(storagePath, photo.Filename))
    
    // Verify DB record
    dbPhoto := queries.GetPhoto(photo.ID)
    assert.Equal(t, photo.ID, dbPhoto.ID)
}
```

## Tests Required
- [ ] Integration test: upload JPEG → WebP conversion
- [ ] Integration test: upload PNG → WebP conversion
- [ ] Integration test: share link with view limit
- [ ] Integration test: share link expiration
- [ ] Integration test: admin login creates session
- [ ] Integration test: album delete cascades to photos

## PR Checklist
- [ ] All tests use temporary resources (no side effects)
- [ ] Test fixtures included in `testdata/`
- [ ] Tests run in parallel where possible
- [ ] Tests pass: `go test ./... -v -race`
- [ ] Code coverage > 60% for critical paths

## Git Workflow
```bash
git checkout -b feat/integration-tests
# Write integration tests
go test ./... -v -race -cover
git add internal/testutil/ testdata/ *_integration_test.go
git commit -m "test: add core integration tests and fixtures"
git push origin feat/integration-tests
# Open PR: "Add integration tests for critical workflows"
```

## Notes
- Use `testing.Short()` to skip slow tests in quick runs
- Test fixtures should be small (< 100KB) to keep repo lightweight
- Consider using `testify/assert` for cleaner assertions
- Run tests with `-race` flag to detect race conditions
