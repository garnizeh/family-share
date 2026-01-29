package pipeline

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"strings"

	"familyshare/internal/db/sqlc"
)

// MaxPipelineDimension is the maximum width/height used when resizing images
// before encoding to WebP for storage.
const MaxPipelineDimension = 1920

// ProcessAndSave runs the full pipeline: validate+decode -> exif -> resize -> encode -> save
func ProcessAndSave(
	ctx context.Context,
	db *sql.DB,
	albumID int64,
	upload io.ReadSeeker,
	maxBytes int64,
	baseDir string,
) (*sqlc.Photo, error) {
	return ProcessAndSaveWithFormat(ctx, db, albumID, upload, maxBytes, baseDir, "webp")
}

// ProcessAndSaveWithFormat runs the full pipeline and encodes to the requested format.
// Supported formats: webp, avif.
func ProcessAndSaveWithFormat(
	ctx context.Context,
	db *sql.DB,
	albumID int64,
	upload io.ReadSeeker,
	maxBytes int64,
	baseDir string,
	format string,
) (*sqlc.Photo, error) {
	// Validate and decode
	img, _, err := ValidateAndDecode(upload, maxBytes)
	if err != nil {
		return nil, fmt.Errorf("validate decode: %w", err)
	}

	// Apply EXIF orientation if available (requires reset of reader)
	if upload != nil {
		if _, err := upload.Seek(0, 0); err == nil {
			img, _ = ApplyEXIFOrientation(img, upload)
		}
	}

	// Resize to pipeline maximum
	img = Resize(img, MaxPipelineDimension)

	format = normalizeFormat(format)
	if format == "" {
		format = "webp"
	}

	// Encode
	var buf bytes.Buffer
	switch format {
	case "avif":
		if err := EncodeAVIF(img, &buf, DefaultAVIFQuality, DefaultAVIFSpeed); err != nil {
			return nil, fmt.Errorf("encode avif: %w", err)
		}
	case "webp":
		if err := EncodeWebP(img, &buf, DefaultWebPQuality); err != nil {
			return nil, fmt.Errorf("encode webp: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	sizeBytes := buf.Len()
	// Save encoded data and create DB record
	_, _, photo, err := SaveProcessedImage(ctx, db, baseDir, albumID, bytes.NewReader(buf.Bytes()), img.Bounds().Dx(), img.Bounds().Dy(), sizeBytes, format)
	if err != nil {
		return nil, fmt.Errorf("save processed image: %w", err)
	}

	return photo, nil
}

func normalizeFormat(format string) string {
	return strings.TrimPrefix(strings.ToLower(format), ".")
}
