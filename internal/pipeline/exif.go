package pipeline

import (
	"image"
	"io"

	"github.com/disintegration/imaging"
	"github.com/rwcarlsen/goexif/exif"
)

// ApplyEXIFOrientation reads EXIF from r (must be an io.ReadSeeker) and applies the
// orientation transform to img. If EXIF is not present or can't be parsed, the
// original img is returned without error.
func ApplyEXIFOrientation(img image.Image, r io.ReadSeeker) (image.Image, error) {
	if r == nil {
		return img, nil
	}

	// Ensure reader is at start
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return img, err
	}

	x, err := exif.Decode(r)
	if err != nil {
		// Not a fatal error for non-JPEGs or images without EXIF
		return img, nil
	}

	tag, err := x.Get(exif.Orientation)
	if err != nil {
		return img, nil
	}
	orient, err := tag.Int(0)
	if err != nil {
		return img, nil
	}

	return orientationTransform(img, orient), nil
}

// orientationTransform applies the necessary flip/rotation for EXIF orientation
// values 1-8. Unknown values return the original image.
func orientationTransform(img image.Image, orientation int) image.Image {
	switch orientation {
	case 1:
		return img
	case 2:
		// Flip horizontal
		return imaging.FlipH(img)
	case 3:
		// Rotate 180
		return imaging.Rotate180(img)
	case 4:
		// Flip vertical
		return imaging.FlipV(img)
	case 5:
		// Transpose: flip horizontal + rotate 90 CCW
		return imaging.Rotate270(imaging.FlipH(img))
	case 6:
		// Rotate 90 CW
		return imaging.Rotate90(img)
	case 7:
		// Transverse: flip horizontal + rotate 90 CW
		return imaging.Rotate90(imaging.FlipH(img))
	case 8:
		// Rotate 90 CCW
		return imaging.Rotate270(img)
	default:
		return img
	}
}
