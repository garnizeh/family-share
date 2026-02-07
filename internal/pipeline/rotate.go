package pipeline

import (
	"fmt"
	"image"
	"os"

	"github.com/chai2010/webp"
	"github.com/disintegration/imaging"
)

// Rotate loads an image from path, rotates it by angle (90, 180, 270),
// and saves it back to the same path. Returns new dimensions and file size.
func Rotate(path string, angle int) (int, int, int64, error) {
	// 1. Open the file
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to open image: %w", err)
	}
	defer f.Close()

	// 2. Decode (assuming WebP since that's what we store)
	img, err := webp.Decode(f)
	if err != nil {
		// Fallback for other formats if necessary, but we enforce WebP
		// Trying generic decode just in case
		f.Seek(0, 0)
		img, _, err = image.Decode(f)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("failed to decode image: %w", err)
		}
	}

	// 3. Rotate
	// imaging.Rotate takes angle in degrees counter-clockwise.
	// 90 (CCW) -> Left
	// -90 (CW) -> Right (or 270)
	// 180 -> Upside down
	rotatedImg := imaging.Rotate(img, float64(angle), image.Transparent)

	// 4. Encode back to WebP
	// We need to write to a temp file first to ensure atomic write or just overwrite safely
	// Since we are overwriting, let's reopen the file for writing or create a new handle
	// But we can't write to the same file while it's open if we were streaming, 
	// but here we have the image in memory.
	// Close input file first explicitly if not effectively closed by defer yet (it will be closed by defer)
	f.Close()

	// Use pipeline.Encode to consistency?
	// But pipeline.Encode returns bytes or writes to writer. 
	// Let's reuse 'Encode' logic if possible, but Encode mainly handles conversion options.
	// For simple rotation, we just want to save it back as WebP 80% quality (standard MVP rule).
	
	outFile, err := os.Create(path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	if err := webp.Encode(outFile, rotatedImg, &webp.Options{Quality: 80}); err != nil {
		return 0, 0, 0, fmt.Errorf("failed to encode webp: %w", err)
	}

	// Get file info for size
	fi, err := outFile.Stat()
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to stat new file: %w", err)
	}

	b := rotatedImg.Bounds()
	return b.Dx(), b.Dy(), fi.Size(), nil
}
