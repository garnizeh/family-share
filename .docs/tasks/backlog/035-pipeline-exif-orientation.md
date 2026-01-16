# Task 035: Image Pipeline — EXIF Orientation Handling

**Milestone:** Storage & Pipeline  
**Points:** 1 (4 hours)  
**Dependencies:** 030  
**Branch:** `feat/pipeline-exif`  
**Labels:** `image-pipeline`, `storage`

## Description
Extract EXIF orientation tag from JPEG images and apply the correct rotation/flip to ensure images display correctly.

## Acceptance Criteria
- [ ] EXIF data extracted from JPEG uploads
- [ ] Orientation tag (1-8) parsed correctly
- [ ] Image rotated/flipped based on orientation value
- [ ] Non-JPEG images skip EXIF processing
- [ ] Corrected image returned for next pipeline stage

## Files to Add/Modify
- `internal/pipeline/exif.go` — EXIF extraction and orientation correction
- `internal/pipeline/exif_test.go` — unit tests with oriented fixtures

## Key Functions
```go
// ApplyEXIFOrientation reads EXIF and corrects image orientation
func ApplyEXIFOrientation(img image.Image, r io.ReadSeeker) (image.Image, error)

// orientationTransform applies the necessary rotation/flip
func orientationTransform(img image.Image, orientation int) image.Image
```

## EXIF Orientation Values
- 1: Normal
- 2: Flip horizontal
- 3: Rotate 180°
- 4: Flip vertical
- 5: Transpose (flip horizontal + rotate 90° CCW)
- 6: Rotate 90° CW
- 7: Transverse (flip horizontal + rotate 90° CW)
- 8: Rotate 90° CCW

## Tests Required
- [ ] Unit test: JPEG with orientation 1 (no change)
- [ ] Unit test: JPEG with orientation 6 (rotate 90° CW)
- [ ] Unit test: JPEG with orientation 8 (rotate 90° CCW)
- [ ] Unit test: PNG without EXIF (no error, return original)
- [ ] Unit test: corrupt EXIF data (graceful fallback)

## PR Checklist
- [ ] All 8 orientation values handled correctly
- [ ] Library `github.com/rwcarlsen/goexif/exif` used
- [ ] Tests include sample images with EXIF tags
- [ ] Non-JPEG formats handled gracefully
- [ ] Tests pass: `go test ./internal/pipeline/... -v`

## Git Workflow
```bash
git checkout -b feat/pipeline-exif
# Implement EXIF handling
go test ./internal/pipeline/... -v -cover
git add internal/pipeline/
git commit -m "feat: add EXIF orientation handling to pipeline"
git push origin feat/pipeline-exif
# Open PR: "Handle EXIF orientation in image pipeline"
```

## Notes
- EXIF is JPEG-specific; skip for PNG/WebP
- Use `io.ReadSeeker` to allow rewinding for EXIF + decode
- Consider caching orientation value to avoid double-read
- Apply orientation before resizing to ensure correct aspect ratio
