package pipeline

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"

	webp "github.com/chai2010/webp"
)

// DetectFormat reads up to 512 bytes from r and returns the detected MIME type.
// Note: this will consume from r.
func DetectFormat(r io.Reader) (string, error) {
	buf := make([]byte, 512)
	n, err := io.ReadAtLeast(r, buf, 1)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return "", err
	}
	return http.DetectContentType(buf[:n]), nil
}

// ValidateAndDecode reads up to maxBytes from r, checks content type, decodes to image.Image
// and validates dimensions (MaxDimension).
func ValidateAndDecode(r io.Reader, maxBytes int64) (image.Image, string, error) {
	// read up to maxBytes+1 to detect overflow
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, "", err
	}
	if int64(len(data)) > maxBytes {
		return nil, "", ErrTooLarge
	}

	// detect content type
	ct := http.DetectContentType(data)

	var img image.Image
	var decodeErr error

	switch {
	case strings.HasPrefix(ct, "image/jpeg"):
		img, decodeErr = jpeg.Decode(bytes.NewReader(data))
	case strings.HasPrefix(ct, "image/png"):
		img, decodeErr = png.Decode(bytes.NewReader(data))
	case strings.HasPrefix(ct, "image/gif"):
		img, _, decodeErr = image.Decode(bytes.NewReader(data))
	case strings.HasPrefix(ct, "image/webp"):
		img, decodeErr = webp.Decode(bytes.NewReader(data))
	default:
		return nil, ct, ErrNotAnImage
	}
	if decodeErr != nil {
		return nil, ct, decodeErr
	}

	// validate dimensions
	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if w <= 0 || h <= 0 || w > MaxDimension || h > MaxDimension {
		return nil, ct, ErrInvalidDimensions
	}

	return img, ct, nil
}
