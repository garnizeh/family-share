# Task 030: Image Pipeline — Decode and Validate

**Milestone:** Storage & Pipeline  
**Points:** 1 (5 hours)  
**Dependencies:** 025  
**Branch:** `feat/pipeline-decode`  
**Labels:** `image-pipeline`, `storage`

## Description
Implement the first stage of the image processing pipeline: decode uploaded images, validate format, and perform content sniffing.

## Acceptance Criteria
- [x] Accept multipart upload stream with size limit
- [x] Detect image MIME type (JPEG, PNG, WebP, HEIC, etc.)
- [x] Reject non-image uploads
- [x] Decode to `image.Image` in memory
- [x] Handle decode errors gracefully
- [x] Validate image dimensions are reasonable (e.g., max 8000x8000)

## Files to Add/Modify
- `internal/pipeline/decode.go` — decode and validation logic
- `internal/pipeline/types.go` — pipeline types and errors
- `internal/pipeline/decode_test.go` — unit tests with fixtures

## Key Functions
```go
// ValidateAndDecode reads upload, validates type, and decodes to image.Image
func ValidateAndDecode(r io.Reader, maxBytes int64) (image.Image, string, error)

// DetectFormat uses content sniffing to determine image type
func DetectFormat(r io.Reader) (string, error)

// Common errors
var (
    ErrNotAnImage = errors.New("uploaded file is not an image")
    ErrTooLarge = errors.New("image exceeds size limit")
    ErrInvalidDimensions = errors.New("image dimensions out of range")
)
```

## Tests Required
- [ ] Unit test: decode valid JPEG
- [ ] Unit test: decode valid PNG
- [ ] Unit test: reject PDF or text file
- [ ] Unit test: reject oversized image (> maxBytes)
- [ ] Unit test: reject image with extreme dimensions (e.g., 20000x1)

## PR Checklist
- [ ] All supported formats tested (JPEG, PNG, WebP)
- [ ] Error messages are user-friendly
- [ ] Memory usage is bounded (streaming where possible)
- [ ] Tests pass: `go test ./internal/pipeline/... -v`
- [ ] Code handles partial reads and EOF correctly

## Git Workflow
```bash
git checkout -b feat/pipeline-decode
# Implement decode logic
go test ./internal/pipeline/... -v -cover
git add internal/pipeline/
git commit -m "feat: implement image decode and validation stage"
git push origin feat/pipeline-decode
# Open PR: "Add image decode and validation to pipeline"
```

## Notes
- Use `image.DecodeConfig` first to check dimensions without full decode
- Limit in-memory buffer size to prevent DoS
- Support common formats; defer HEIC/AVIF decode to post-MVP if needed
- Use `http.DetectContentType` for MIME sniffing
