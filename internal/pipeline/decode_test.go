package pipeline

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"testing"

	"github.com/gen2brain/avif"
)

func encodeJPEG(w io.Writer) error {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	return jpeg.Encode(w, img, &jpeg.Options{Quality: 80})
}

func encodePNG(w io.Writer) error {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	return png.Encode(w, img)
}

func TestValidateAndDecodeJPEG(t *testing.T) {
	var b bytes.Buffer
	if err := encodeJPEG(&b); err != nil {
		t.Fatal(err)
	}
	img, ct, err := ValidateAndDecode(&b, 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == nil {
		t.Fatalf("expected image, got nil")
	}
	if ct == "" {
		t.Fatalf("expected content type")
	}
}

func TestValidateAndDecodePNG(t *testing.T) {
	var b bytes.Buffer
	if err := encodePNG(&b); err != nil {
		t.Fatal(err)
	}
	img, _, err := ValidateAndDecode(&b, 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if img == nil {
		t.Fatalf("expected image, got nil")
	}
}

func TestValidateAndDecodeAVIF(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	var b bytes.Buffer
	if err := avif.Encode(&b, img, avif.Options{Quality: 60, Speed: 6}); err != nil {
		t.Fatalf("encode avif: %v", err)
	}
	decoded, ct, err := ValidateAndDecode(bytes.NewReader(b.Bytes()), 1<<20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decoded == nil {
		t.Fatalf("expected image, got nil")
	}
	if ct != "image/avif" {
		t.Fatalf("expected image/avif, got %s", ct)
	}
}

func TestRejectText(t *testing.T) {
	b := bytes.NewBufferString("this is not an image")
	_, _, err := ValidateAndDecode(b, 1024)
	if err == nil {
		t.Fatalf("expected error for non-image")
	}
}

func TestRejectTooLarge(t *testing.T) {
	// create a reader larger than limit
	data := make([]byte, 1024*10)
	for i := range data {
		data[i] = 'a'
	}
	_, _, err := ValidateAndDecode(bytes.NewReader(data), 1024)
	if err != ErrTooLarge {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestRejectInvalidDimensions(t *testing.T) {
	// create a very wide image
	img := image.NewRGBA(image.Rect(0, 0, MaxDimension+1, 1))
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 80}); err != nil {
		t.Fatal(err)
	}
	_, _, err := ValidateAndDecode(&b, 1<<20)
	if err != ErrInvalidDimensions {
		t.Fatalf("expected ErrInvalidDimensions, got %v", err)
	}
}
