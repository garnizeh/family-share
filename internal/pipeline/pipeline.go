package pipeline

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"

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

	// Encode to WebP
	var buf bytes.Buffer
	if err := EncodeWebP(img, &buf, DefaultWebPQuality); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}

	sizeBytes := buf.Len()
	// Save encoded data and create DB record
	photoID, _, err := SaveProcessedImage(ctx, db, albumID, bytes.NewReader(buf.Bytes()), img.Bounds().Dx(), img.Bounds().Dy(), sizeBytes, "webp")
	if err != nil {
		return nil, fmt.Errorf("save processed image: %w", err)
	}

	q := sqlc.New(db)
	p, err := q.GetPhoto(ctx, photoID)
	if err != nil {
		return nil, fmt.Errorf("get photo: %w", err)
	}
	return &p, nil
}
