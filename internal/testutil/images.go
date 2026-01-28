package testutil

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
)

// GenerateSampleImages creates sample test images in testdata/images.
// This is a helper for setting up test fixtures.
func GenerateSampleImages() error {
	baseDir := "testdata/images"
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return err
	}

	// Generate sample JPEG
	jpegImg := createGradientImage(800, 600)
	jpegPath := filepath.Join(baseDir, "sample.jpg")
	if err := saveJPEG(jpegPath, jpegImg); err != nil {
		return err
	}

	// Generate sample PNG
	pngImg := createGradientImage(640, 480)
	pngPath := filepath.Join(baseDir, "sample.png")
	if err := savePNG(pngPath, pngImg); err != nil {
		return err
	}

	// Generate large JPEG for size testing
	largeImg := createGradientImage(3000, 2000)
	largePath := filepath.Join(baseDir, "large.jpg")
	if err := saveJPEG(largePath, largeImg); err != nil {
		return err
	}

	return nil
}

func createGradientImage(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	return img
}

func saveJPEG(path string, img image.Image) error {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func savePNG(path string, img image.Image) error {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

func init() {
	// Auto-generate sample images if they don't exist
	if _, err := os.Stat("testdata/images/sample.jpg"); os.IsNotExist(err) {
		GenerateSampleImages()
	}
}
