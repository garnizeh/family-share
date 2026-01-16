package pipeline

import "errors"

var (
	ErrNotAnImage        = errors.New("uploaded file is not an image")
	ErrTooLarge          = errors.New("image exceeds size limit")
	ErrInvalidDimensions = errors.New("image dimensions out of range")
)

// Default maximum dimension (width or height) allowed by validator.
const MaxDimension = 8000
