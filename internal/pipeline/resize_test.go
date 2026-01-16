package pipeline

import (
	"image"
	"testing"
)

func newRGBA(w, h int) image.Image {
	return image.NewRGBA(image.Rect(0, 0, w, h))
}

func TestCalculateDimensions_Landscape(t *testing.T) {
	nw, nh := calculateDimensions(4000, 3000, 1920)
	if nw != 1920 || nh != 1440 {
		t.Fatalf("expected 1920x1440, got %dx%d", nw, nh)
	}
}

func TestCalculateDimensions_Portrait(t *testing.T) {
	nw, nh := calculateDimensions(3000, 4000, 1920)
	if nw != 1440 || nh != 1920 {
		t.Fatalf("expected 1440x1920, got %dx%d", nw, nh)
	}
}

func TestResize_NoUpscale(t *testing.T) {
	img := newRGBA(1000, 800)
	out := Resize(img, 1920)
	if out.Bounds().Dx() != 1000 || out.Bounds().Dy() != 800 {
		t.Fatalf("expected unchanged 1000x800, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}

func TestResize_ExactFit(t *testing.T) {
	img := newRGBA(1920, 1080)
	out := Resize(img, 1920)
	if out.Bounds().Dx() != 1920 || out.Bounds().Dy() != 1080 {
		t.Fatalf("expected unchanged 1920x1080, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}

func TestResize_LargeLandscape(t *testing.T) {
	img := newRGBA(4000, 3000)
	out := Resize(img, 1920)
	if out.Bounds().Dx() != 1920 || out.Bounds().Dy() != 1440 {
		t.Fatalf("expected 1920x1440, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}

func TestResize_LargePortrait(t *testing.T) {
	img := newRGBA(3000, 4000)
	out := Resize(img, 1920)
	if out.Bounds().Dx() != 1440 || out.Bounds().Dy() != 1920 {
		t.Fatalf("expected 1440x1920, got %dx%d", out.Bounds().Dx(), out.Bounds().Dy())
	}
}
