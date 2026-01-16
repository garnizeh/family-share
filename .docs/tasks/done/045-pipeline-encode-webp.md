# Task 045: Image Pipeline — Encode to WebP

**Milestone:** Storage & Pipeline  
**Points:** 1 (5 hours)  
**Dependencies:** 040  
**Branch:** `feat/pipeline-webp`  
**Labels:** `image-pipeline`, `storage`

## Description
Encode processed images to WebP format at 80% quality to minimize storage usage while maintaining acceptable visual quality.

## Acceptance Criteria
- [x] `image.Image` encoded to WebP bytes
- [x] Quality setting at 80 (configurable)
- [x] Encoded data written to `io.Writer`
- [x] Encoding errors handled gracefully
- [x] Final size logged for monitoring

## Files to Add/Modify
- `internal/pipeline/encode.go` — WebP encoding logic
- `internal/pipeline/encode_test.go` — unit tests

## Key Functions
```go
// EncodeWebP encodes image to WebP format
func EncodeWebP(img image.Image, w io.Writer, quality int) error

// DefaultWebPQuality is the standard quality setting
const DefaultWebPQuality = 80
```

## Tests Required
- [ ] Unit test: encode valid image to WebP
- [ ] Unit test: encoded output is valid WebP (decode and verify)
- [ ] Unit test: quality setting affects output size
- [ ] Unit test: encoding to closed writer returns error
- [ ] Integration test: full pipeline decode → resize → encode

## PR Checklist
- [ ] Library `github.com/chai2010/webp` used
- [ ] Quality setting is configurable
- [ ] Encoded images are visually acceptable (manual spot-check)
- [ ] Tests pass: `go test ./internal/pipeline/... -v`
- [ ] Memory usage is reasonable (no large buffer leaks)

## Git Workflow
```bash
git checkout -b feat/pipeline-webp
# Implement WebP encoding
go test ./internal/pipeline/... -v -cover
git add internal/pipeline/
git commit -m "feat: add WebP encoding to image pipeline"
git push origin feat/pipeline-webp
# Open PR: "Implement WebP encoding for processed images"
```

## Notes
- WebP quality 80 is good balance of size vs quality
- Consider logging compression ratio (original vs WebP size)
- Handle grayscale and RGBA images correctly
- Test with photos containing transparency (if supported)
