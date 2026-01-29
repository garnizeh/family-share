package storage

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PhotoPath returns the storage path for a photo using layout:
// {baseDir}/photos/{yyyy}/{mm}/{album_id}/{photo_id}.{ext}
// It uses the current time and should be avoided for persisted photo paths.
func PhotoPath(baseDir string, albumID, photoID int64, format string) string {
	return PhotoPathAt(baseDir, albumID, photoID, format, time.Now().UTC())
}

// PhotoPathAt returns the storage path for a photo using a provided timestamp.
// This is the stable path used for persisted photos.
func PhotoPathAt(baseDir string, albumID, photoID int64, format string, createdAt time.Time) string {
	t := createdAt.UTC()
	year := t.Format("2006")
	month := t.Format("01")
	ext := strings.ToLower(strings.TrimPrefix(format, "."))
	return filepath.Join(baseDir, "photos", year, month, strconv.FormatInt(albumID, 10), fmt.Sprintf("%d.%s", photoID, ext))
}

// ThumbnailPath returns a thumbnail path next to the original with a _thumb suffix.
// It uses the current time and should be avoided for persisted photo paths.
func ThumbnailPath(baseDir string, albumID, photoID int64) string {
	return ThumbnailPathAt(baseDir, albumID, photoID, time.Now().UTC())
}

// ThumbnailPathAt returns a thumbnail path next to the original with a _thumb suffix
// using the provided timestamp.
func ThumbnailPathAt(baseDir string, albumID, photoID int64, createdAt time.Time) string {
	// use webp thumbnails by default
	p := PhotoPathAt(baseDir, albumID, photoID, "webp", createdAt)
	ext := filepath.Ext(p)
	without := strings.TrimSuffix(p, ext)
	return fmt.Sprintf("%s_thumb%s", without, ext)
}
