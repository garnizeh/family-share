package testutil

import (
	"bytes"
	"context"
	"database/sql"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"familyshare/internal/db/sqlc"

	"golang.org/x/crypto/bcrypt"
)

// CreateTestAlbum creates a test album in the database.
func CreateTestAlbum(t *testing.T, q *sqlc.Queries, title, description string) *sqlc.Album {
	t.Helper()

	params := sqlc.CreateAlbumParams{
		Title:       title,
		Description: sql.NullString{String: description, Valid: description != ""},
	}

	album, err := q.CreateAlbum(context.Background(), params)
	if err != nil {
		t.Fatalf("failed to create test album: %v", err)
	}

	return &album
}

// CreateTestPhoto creates a test photo record in the database.
func CreateTestPhoto(t *testing.T, q *sqlc.Queries, albumID int64, filename string) *sqlc.Photo {
	t.Helper()

	params := sqlc.CreatePhotoParams{
		AlbumID:   albumID,
		Filename:  filename,
		Width:     1920,
		Height:    1080,
		SizeBytes: 150000,
		Format:    "webp",
	}

	photo, err := q.CreatePhoto(context.Background(), params)
	if err != nil {
		t.Fatalf("failed to create test photo: %v", err)
	}

	return &photo
}

// CreateTestShareLink creates a test share link in the database.
func CreateTestShareLink(t *testing.T, q *sqlc.Queries, albumID int64, token string, maxViews int64, expiresAt time.Time) *sqlc.ShareLink {
	t.Helper()

	params := sqlc.CreateShareLinkParams{
		Token:      token,
		TargetType: "album",
		TargetID:   albumID,
		MaxViews: sql.NullInt64{
			Int64: maxViews,
			Valid: maxViews > 0,
		},
		ExpiresAt: sql.NullTime{
			Time:  expiresAt.UTC(),
			Valid: !expiresAt.IsZero(),
		},
	}

	link, err := q.CreateShareLink(context.Background(), params)
	if err != nil {
		t.Fatalf("failed to create test share link: %v", err)
	}

	return &link
}

// CreateTestSession creates a test admin session in the database.
func CreateTestSession(t *testing.T, q *sqlc.Queries, token string, expiresAt time.Time) *sqlc.Session {
	t.Helper()

	params := sqlc.CreateSessionParams{
		ID:        token,
		UserID:    "admin",
		ExpiresAt: expiresAt.UTC(),
	}

	session, err := q.CreateSession(context.Background(), params)
	if err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}

	return &session
}

// HashPassword creates a bcrypt hash for testing admin authentication.
func HashPassword(t *testing.T, password string) string {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	return string(hash)
}

// GenerateTestImage creates a simple test image with the specified format.
// Returns an io.ReadSeeker containing the encoded image.
func GenerateTestImage(t *testing.T, format string, width, height int) io.ReadSeeker {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with a gradient pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}

	var buf bytes.Buffer

	switch format {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
			t.Fatalf("failed to encode JPEG: %v", err)
		}
	case "png":
		if err := png.Encode(&buf, img); err != nil {
			t.Fatalf("failed to encode PNG: %v", err)
		}
	default:
		t.Fatalf("unsupported image format: %s", format)
	}

	return bytes.NewReader(buf.Bytes())
}

// LoadTestImage loads a test image from the testdata directory.
// If the file doesn't exist, it generates one and saves it.
func LoadTestImage(t *testing.T, name string) io.Reader {
	t.Helper()

	path := filepath.Join("testdata", "images", name)

	// Try to load existing file
	if data, err := os.ReadFile(path); err == nil {
		return bytes.NewReader(data)
	}

	// Generate image based on extension
	var format string
	var width, height int

	switch filepath.Ext(name) {
	case ".jpg", ".jpeg":
		format = "jpeg"
		width, height = 800, 600
	case ".png":
		format = "png"
		width, height = 800, 600
	default:
		t.Fatalf("unknown image extension for %s", name)
	}

	img := GenerateTestImage(t, format, width, height)

	// Save for future use
	if err := os.MkdirAll(filepath.Dir(path), 0755); err == nil {
		data, _ := io.ReadAll(img)
		_ = os.WriteFile(path, data, 0644)
		return bytes.NewReader(data)
	}

	return img
}
