package pipeline

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"testing"

	webp "github.com/chai2010/webp"
	"github.com/gen2brain/avif"
)

func smallTestImage() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 64, 48))
	// put a red dot to avoid fully blank image optimizations
	img.Set(1, 1, color.RGBA{255, 0, 0, 255})
	return img
}

func TestEncodeWebP_ValidImage(t *testing.T) {
	img := smallTestImage()
	var buf bytes.Buffer
	if err := EncodeWebP(img, &buf, DefaultWebPQuality); err != nil {
		t.Fatalf("EncodeWebP failed: %v", err)
	}
	// ensure output decodes as WebP
	if _, err := webp.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("decoded webp failed: %v", err)
	}
}

func TestEncodeWebP_QualityAffectsSize(t *testing.T) {
	img := smallTestImage()
	var low bytes.Buffer
	var high bytes.Buffer
	if err := EncodeWebP(img, &low, 30); err != nil {
		t.Fatalf("encode low quality failed: %v", err)
	}
	if err := EncodeWebP(img, &high, 90); err != nil {
		t.Fatalf("encode high quality failed: %v", err)
	}
	if low.Len() == 0 || high.Len() == 0 {
		t.Fatalf("encoded output empty")
	}
	if low.Len() >= high.Len() {
		t.Fatalf("expected low quality size < high quality size, got %d >= %d", low.Len(), high.Len())
	}
}

type badWriter struct{}

func (badWriter) Write(p []byte) (int, error) { return 0, fmt.Errorf("closed writer") }

func TestEncodeWebP_ClosedWriter(t *testing.T) {
	img := smallTestImage()
	var bw badWriter
	if err := EncodeWebP(img, bw, DefaultWebPQuality); err == nil {
		t.Fatalf("expected error when writing to closed writer")
	}
}

func TestEncodeWebP_IntegrationResizeEncodeDecode(t *testing.T) {
	// create a larger image, resize and encode, then decode
	img := image.NewRGBA(image.Rect(0, 0, 3000, 2000))
	// mark a pixel so decoding can be validated roughly by dimensions
	img.Set(2, 2, color.RGBA{1, 2, 3, 255})
	resized := Resize(img, 1920)
	var buf bytes.Buffer
	if err := EncodeWebP(resized, &buf, DefaultWebPQuality); err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	// decode and verify dimensions roughly match resized bounds
	out, err := webp.Decode(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("webp decode failed: %v", err)
	}
	if out.Bounds().Dx() != resized.Bounds().Dx() || out.Bounds().Dy() != resized.Bounds().Dy() {
		t.Fatalf("decoded dims %vx%v don't match resized %vx%v", out.Bounds().Dx(), out.Bounds().Dy(), resized.Bounds().Dx(), resized.Bounds().Dy())
	}
}

func TestEncodeAVIF_ValidImage(t *testing.T) {
	img := smallTestImage()
	var buf bytes.Buffer
	if err := EncodeAVIF(img, &buf, DefaultAVIFQuality, DefaultAVIFSpeed); err != nil {
		t.Fatalf("EncodeAVIF failed: %v", err)
	}
	if buf.Len() == 0 {
		t.Fatalf("encoded output empty")
	}
	if _, err := avif.Decode(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("decoded avif failed: %v", err)
	}
}
