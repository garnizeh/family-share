package storage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestPhotoPathAndThumbnail(t *testing.T) {
	tmp := t.TempDir()
	createdAt := time.Date(2025, time.December, 5, 12, 0, 0, 0, time.UTC)
	p := PhotoPathAt(tmp, 42, 7, "webp", createdAt)
	if !strings.Contains(p, "photos") {
		t.Fatalf("expected path to contain photos: %s", p)
	}
	if !strings.Contains(p, filepath.Join("photos", "2025", "12")) {
		t.Fatalf("expected path to include year/month from created_at: %s", p)
	}
	if !strings.Contains(p, strconv.FormatInt(42, 10)) {
		t.Fatalf("expected album id in path: %s", p)
	}
	if !strings.HasSuffix(p, "7.webp") {
		t.Fatalf("expected suffix 7.webp got: %s", p)
	}

	th := ThumbnailPathAt(tmp, 42, 7, createdAt)
	if !strings.HasSuffix(th, "_thumb.webp") {
		t.Fatalf("expected thumbnail suffix _thumb.webp got: %s", th)
	}
}

func TestEnsureDir(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "photos", "2026", "01", "42")
	if err := EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir error: %v", err)
	}
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("expected dir, got file")
	}
}

func TestAtomicWriteSuccess(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "file.bin")
	data := bytes.NewReader([]byte("hello world"))
	if err := AtomicWrite(path, data); err != nil {
		t.Fatalf("AtomicWrite failed: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(b) != "hello world" {
		t.Fatalf("unexpected contents: %s", string(b))
	}
}

type failReader struct{ n int }

func (f *failReader) Read(p []byte) (int, error) {
	if f.n <= 0 {
		return 0, io.ErrUnexpectedEOF
	}
	// write one byte then fail
	p[0] = 'x'
	f.n--
	return 1, nil
}

func TestAtomicWritePartialFailure(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "partial.bin")
	fr := &failReader{n: 0}
	if err := AtomicWrite(path, fr); err == nil {
		t.Fatalf("expected error from AtomicWrite with failing reader")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected no final file on failure, got: %v", err)
	}
}

func TestCleanupExecute(t *testing.T) {
	tmp := t.TempDir()
	f1 := filepath.Join(tmp, "a.tmp")
	f2 := filepath.Join(tmp, "b.tmp")
	if err := os.WriteFile(f1, []byte("x"), 0o600); err != nil {
		t.Fatalf("write f1: %v", err)
	}
	if err := os.WriteFile(f2, []byte("y"), 0o600); err != nil {
		t.Fatalf("write f2: %v", err)
	}
	var c Cleanup
	c.Add(f1)
	c.Add(f2)
	if err := c.Execute(); err != nil {
		t.Fatalf("cleanup error: %v", err)
	}
	if _, err := os.Stat(f1); !os.IsNotExist(err) {
		t.Fatalf("f1 should be removed")
	}
	if _, err := os.Stat(f2); !os.IsNotExist(err) {
		t.Fatalf("f2 should be removed")
	}
}
