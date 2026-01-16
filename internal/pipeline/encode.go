package pipeline

import (
	"errors"
	"image"
	"io"
	"log"

	webp "github.com/chai2010/webp"
)

// DefaultWebPQuality is the standard quality used for lossy WebP encoding.
const DefaultWebPQuality = 80

// EncodeWebP encodes img to WebP written to w with given quality (0-100).
// It logs the final encoded size. Returns an error from the encoder or writer.
func EncodeWebP(img image.Image, w io.Writer, quality int) error {
	if img == nil {
		return errors.New("nil image")
	}
	if w == nil {
		return errors.New("nil writer")
	}
	if quality < 0 {
		quality = 0
	}
	if quality > 100 {
		quality = 100
	}

	// counting writer to capture encoded size
	c := &countingWriter{w: w}
	opts := &webp.Options{Quality: float32(quality)}
	if err := webp.Encode(c, img, opts); err != nil {
		return err
	}

	log.Printf("webp encoded size=%d quality=%d", c.n, quality)
	return nil
}

// countingWriter wraps an io.Writer and counts bytes written.
type countingWriter struct {
	w io.Writer
	n int64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	m, err := c.w.Write(p)
	c.n += int64(m)
	return m, err
}
