package pipeline

import (
	"image"

	"github.com/disintegration/imaging"
)

// Resize reduces image to fit within maxDimension, preserving aspect ratio.
// Does not upscale smaller images.
func Resize(img image.Image, maxDimension int) image.Image {
    if img == nil {
        return nil
    }
    b := img.Bounds()
    w := b.Dx()
    h := b.Dy()
    if w <= 0 || h <= 0 {
        return img
    }

    if w <= maxDimension && h <= maxDimension {
        return img
    }

    nw, nh := calculateDimensions(w, h, maxDimension)
    // imaging.Resize will perform high-quality resampling; use Lanczos filter.
    return imaging.Resize(img, nw, nh, imaging.Lanczos)
}

// calculateDimensions computes new width/height preserving aspect ratio and
// ensuring the larger dimension equals maxDim (unless no resize needed).
func calculateDimensions(origWidth, origHeight, maxDim int) (int, int) {
    if origWidth <= 0 || origHeight <= 0 || maxDim <= 0 {
        return origWidth, origHeight
    }
    if origWidth <= maxDim && origHeight <= maxDim {
        return origWidth, origHeight
    }
    if origWidth > origHeight {
        newW := maxDim
        newH := (origHeight * maxDim) / origWidth
        if newH < 1 {
            newH = 1
        }
        return newW, newH
    }
    newH := maxDim
    newW := (origWidth * maxDim) / origHeight
    if newW < 1 {
        newW = 1
    }
    return newW, newH
}
