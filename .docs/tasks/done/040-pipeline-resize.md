# Task 040: Image Pipeline — Resize to Max Dimensions

**Milestone:** Storage & Pipeline  
**Points:** 1 (4 hours)  
**Dependencies:** 035  
**Branch:** `feat/pipeline-resize`  
**Labels:** `image-pipeline`, `storage`

## Description
Resize images to a maximum dimension (1920px width or height) while preserving aspect ratio. Skip resizing if image is already smaller.

## Acceptance Criteria
- [x] Images larger than 1920px (either dimension) are resized
- [x] Aspect ratio is preserved
- [x] Images smaller than 1920px are not upscaled
- [x] High-quality resampling filter used (Lanczos)
- [x] Resized image returned as `image.Image`

## Files to Add/Modify
- `internal/pipeline/resize.go` — resize logic
- `internal/pipeline/resize_test.go` — unit tests

## Key Functions
```go
// Resize reduces image to fit within maxDimension, preserving aspect ratio
func Resize(img image.Image, maxDimension int) image.Image

// calculateDimensions computes new width/height
func calculateDimensions(origWidth, origHeight, maxDim int) (int, int)
```

## Resize Logic
```
if width <= maxDim && height <= maxDim:
    return original (no resize)

if width > height:
    newWidth = maxDim
    newHeight = (height * maxDim) / width
else:
    newHeight = maxDim
    newWidth = (width * maxDim) / height
```

## Tests Required
- [ ] Unit test: 4000x3000 image → 1920x1440
- [ ] Unit test: 3000x4000 image → 1440x1920
- [ ] Unit test: 1000x800 image → unchanged (no upscale)
- [ ] Unit test: 1920x1080 image → unchanged (exact fit)
- [ ] Unit test: aspect ratio preserved in all cases

## PR Checklist
- [ ] Library `github.com/disintegration/imaging` used
- [ ] Lanczos resampling filter configured
- [ ] No upscaling occurs
- [ ] Tests pass: `go test ./internal/pipeline/... -v`
- [ ] Visual quality verified with sample images

## Git Workflow
```bash
git checkout -b feat/pipeline-resize
# Implement resize logic
go test ./internal/pipeline/... -v -cover
git add internal/pipeline/
git commit -m "feat: add image resizing to pipeline"
git push origin feat/pipeline-resize
# Open PR: "Implement image resizing with aspect ratio preservation"
```

## Notes
- Use `imaging.Resize` with `imaging.Lanczos` filter
- Avoid multiple resize passes (quality degradation)
- Consider making maxDimension configurable via env var
- Document memory usage for large images
