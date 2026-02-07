package pipeline

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/chai2010/webp"
)

func createTestWebP(t *testing.T, dir string, w, h int) string {
	path := filepath.Join(dir, "test_rotate.webp")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	defer f.Close()

	// Create a simple image with distinct colors to verify rotation
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	// Top-left red
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})
	// Bottom-right blue
	img.Set(w-1, h-1, color.RGBA{0, 0, 255, 255})

	if err := webp.Encode(f, img, &webp.Options{Quality: 80}); err != nil {
		t.Fatalf("failed to encode webp: %v", err)
	}
	return path
}

func TestRotate(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		angle      int
		origW      int
		origH      int
		expectW    int
		expectH    int
		expectSize bool // Check if size > 0
	}{
		{
			name:       "Rotate 90",
			angle:      90,
			origW:      100,
			origH:      50,
			expectW:    50,
			expectH:    100,
			expectSize: true,
		},
		{
			name:       "Rotate -90",
			angle:      -90,
			origW:      100,
			origH:      50,
			expectW:    50,
			expectH:    100,
			expectSize: true,
		},
		{
			name:       "Rotate 180",
			angle:      180,
			origW:      100,
			origH:      50,
			expectW:    100,
			expectH:    50,
			expectSize: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := createTestWebP(t, tempDir, tt.origW, tt.origH)

			w, h, size, err := Rotate(path, tt.angle)
			if err != nil {
				t.Fatalf("Rotate() error = %v", err)
			}

			if w != tt.expectW {
				t.Errorf("Rotate() width = %v, want %v", w, tt.expectW)
			}
			if h != tt.expectH {
				t.Errorf("Rotate() height = %v, want %v", h, tt.expectH)
			}
			if tt.expectSize && size <= 0 {
				t.Errorf("Rotate() size = %v, want > 0", size)
			}

			// Verify dimensions from file content
			f, err := os.Open(path)
			if err != nil {
				t.Fatalf("failed to open rotated file: %v", err)
			}
			defer f.Close()
			cfg, err := webp.DecodeConfig(f)
			if err != nil {
				t.Fatalf("failed to decode config: %v", err)
			}
			if cfg.Width != tt.expectW || cfg.Height != tt.expectH {
				t.Errorf("file dimensions = %dx%d, want %dx%d", cfg.Width, cfg.Height, tt.expectW, tt.expectH)
			}
		})
	}
}

func TestRotate_InvalidFile(t *testing.T) {
	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "nonexistent.webp")

	_, _, _, err := Rotate(path, 90)
	if err == nil {
		t.Error("Rotate() expected error for nonexistent file, got nil")
	}
}
