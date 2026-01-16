package pipeline

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"testing"
)

// Helper to build a simple image with a colored pixel to track transforms.
func coloredImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// fill transparent
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{0, 0, 0, 0})
		}
	}
	// put a red pixel at (1,0)
	if w > 1 && h > 0 {
		img.Set(1, 0, color.RGBA{255, 0, 0, 255})
	}
	return img
}

func TestOrientationTransform_Basic(t *testing.T) {
	src := coloredImage(3, 2) // width 3, height 2

	// orientation 1 -> unchanged
	o1 := orientationTransform(src, 1)
	if o1.Bounds().Dx() != 3 || o1.Bounds().Dy() != 2 {
		t.Fatalf("orientation 1 should preserve bounds")
	}

	// orientation 6 -> rotate 90 CW -> bounds swapped
	o6 := orientationTransform(src, 6)
	if o6.Bounds().Dx() != 2 || o6.Bounds().Dy() != 3 {
		t.Fatalf("orientation 6 should swap width/height")
	}

	// orientation 8 -> rotate 90 CCW -> bounds swapped
	o8 := orientationTransform(src, 8)
	if o8.Bounds().Dx() != 2 || o8.Bounds().Dy() != 3 {
		t.Fatalf("orientation 8 should swap width/height")
	}

	// orientation 3 -> rotate 180 -> bounds same
	o3 := orientationTransform(src, 3)
	if o3.Bounds().Dx() != 3 || o3.Bounds().Dy() != 2 {
		t.Fatalf("orientation 3 should preserve bounds")
	}
}

func TestApplyEXIFOrientation_NonJPEGOrNoEXIF(t *testing.T) {
	// PNG (no EXIF) should return original image without error
	buf := &bytes.Buffer{}
	img := coloredImage(4, 3)
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}

	// Use a ReadSeeker
	rs := bytes.NewReader(buf.Bytes())
	out, err := ApplyEXIFOrientation(img, rs)
	if err != nil {
		t.Fatalf("ApplyEXIFOrientation returned error for PNG: %v", err)
	}
	if out.Bounds() != img.Bounds() {
		t.Fatalf("expected bounds unchanged for PNG/no-exif")
	}
}

func TestApplyEXIFOrientation_JPEG_NoEXIF(t *testing.T) {
	// JPEG without EXIF should also be a no-op
	buf := &bytes.Buffer{}
	img := coloredImage(5, 4)
	if err := jpeg.Encode(buf, img, nil); err != nil {
		t.Fatalf("jpeg encode: %v", err)
	}
	rs := bytes.NewReader(buf.Bytes())
	out, err := ApplyEXIFOrientation(img, rs)
	if err != nil {
		t.Fatalf("ApplyEXIFOrientation returned error for JPEG no-exif: %v", err)
	}
	if out.Bounds() != img.Bounds() {
		t.Fatalf("expected bounds unchanged for JPEG without EXIF")
	}
}

// We can't easily craft a full JPEG with EXIF orientation tag here without
// embedding fixtures; orientationTransform is unit-tested above. This test
// exercises that ApplyEXIFOrientation tolerates corrupt readers gracefully.
func TestApplyEXIFOrientation_CorruptReader(t *testing.T) {
	r := io.NopCloser(bytes.NewReader([]byte("not a valid image")))
	// convert to ReadSeeker
	bs := bytes.NewReader([]byte("not a valid image"))
	img := coloredImage(2, 2)
	out, err := ApplyEXIFOrientation(img, bs)
	if err != nil {
		t.Fatalf("expected no error for corrupt exif decode: %v", err)
	}
	if out.Bounds() != img.Bounds() {
		t.Fatalf("expected original image returned on error")
	}
	_ = r
}
