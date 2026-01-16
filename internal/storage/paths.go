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
func PhotoPath(baseDir string, albumID, photoID int64, format string) string {
	t := time.Now()
	year := t.Format("2006")
	month := t.Format("01")
	ext := strings.ToLower(strings.TrimPrefix(format, "."))
	return filepath.Join(baseDir, "photos", year, month, strconv.FormatInt(albumID, 10), fmt.Sprintf("%d.%s", photoID, ext))
}

// ThumbnailPath returns a thumbnail path next to the original with a _thumb suffix.
func ThumbnailPath(baseDir string, albumID, photoID int64) string {
	// use webp thumbnails by default
	p := PhotoPath(baseDir, albumID, photoID, "webp")
	ext := filepath.Ext(p)
	without := strings.TrimSuffix(p, ext)
	return fmt.Sprintf("%s_thumb%s", without, ext)
}
