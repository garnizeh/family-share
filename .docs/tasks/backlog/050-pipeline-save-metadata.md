# Task 050: Image Pipeline — Atomic Save and Metadata Write

**Milestone:** Storage & Pipeline  
**Points:** 2 (6 hours)  
**Dependencies:** 045, 020  
**Branch:** `feat/pipeline-save`  
**Labels:** `image-pipeline`, `storage`, `database`

## Description
Complete the image pipeline by atomically saving the encoded WebP to disk and writing metadata to the database. Ensure cleanup on failure.

## Acceptance Criteria
- [ ] Encoded WebP saved to correct path using atomic write
- [ ] Photo metadata inserted into database (filename, dimensions, size, format)
- [ ] Transaction used to ensure atomicity (file + DB)
- [ ] On failure, temp files cleaned up and no DB record created
- [ ] Success returns photo ID and storage path

## Files to Add/Modify
- `internal/pipeline/save.go` — save and metadata write logic
- `internal/pipeline/pipeline.go` — orchestrates full pipeline
- `internal/pipeline/pipeline_test.go` — integration tests

## Key Functions
```go
// SaveProcessedImage saves encoded image and creates DB record
func SaveProcessedImage(
    ctx context.Context,
    db *sql.DB,
    albumID int64,
    encodedData io.Reader,
    width, height, sizeBytes int,
    format string,
) (photoID int64, path string, error)

// ProcessAndSave runs full pipeline: decode → EXIF → resize → encode → save
func ProcessAndSave(
    ctx context.Context,
    db *sql.DB,
    albumID int64,
    upload io.ReadSeeker,
    maxBytes int64,
) (*Photo, error)
```

## Tests Required
- [ ] Integration test: full pipeline creates photo and file on disk
- [ ] Integration test: pipeline failure cleans up temp files
- [ ] Integration test: database rollback on file write failure
- [ ] Unit test: metadata extraction (width, height, size)

## PR Checklist
- [ ] Atomic write used via storage helpers (task 025)
- [ ] sqlc `CreatePhoto` query used for DB insert
- [ ] Transaction wraps file write + DB insert
- [ ] Cleanup runs on all error paths (defer)
- [ ] Tests pass: `go test ./internal/pipeline/... -v`
- [ ] Integration test uses temp directory and temp DB

## Git Workflow
```bash
git checkout -b feat/pipeline-save
# Implement save logic and full pipeline orchestration
go test ./internal/pipeline/... -v -cover
git add internal/pipeline/
git commit -m "feat: complete image pipeline with atomic save and metadata"
git push origin feat/pipeline-save
# Open PR: "Complete image processing pipeline with save and metadata"
```

## Notes
- Use deferred cleanup to ensure temp files removed on panic
- Consider returning photo dimensions for display without DB query
- Log each pipeline stage for debugging
- Ensure context cancellation stops processing
