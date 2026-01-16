package pipeline

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"familyshare/internal/db/sqlc"
	"familyshare/internal/storage"
)

// SaveProcessedImage saves encodedData to disk atomically and inserts a photo
// metadata row inside a DB transaction. Returns the created photo ID and the
// final storage path on success.
func SaveProcessedImage(
    ctx context.Context,
    db *sql.DB,
    albumID int64,
    encodedData io.Reader,
    width, height, sizeBytes int,
    format string,
) (int64, string, error) {
    // begin transaction
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return 0, "", fmt.Errorf("begin tx: %w", err)
    }
    // ensure rollback on any early return
    defer func() {
        // if still pending, rollback (ignore error)
        _ = tx.Rollback()
    }()

    q := sqlc.New(tx)

    // generate a simple stored filename (not the full path) for record keeping
    // use timestamp + ext
    ext := strings.TrimPrefix(format, ".")
    filename := fmt.Sprintf("%d.%s", time.Now().UnixNano(), ext)

    p, err := q.CreatePhoto(ctx, sqlc.CreatePhotoParams{
        AlbumID:   albumID,
        Filename:  filename,
        Width:     int64(width),
        Height:    int64(height),
        SizeBytes: int64(sizeBytes),
        Format:    ext,
    })
    if err != nil {
        return 0, "", fmt.Errorf("create photo record: %w", err)
    }

    // determine storage path using env-configured base dir
    base := os.Getenv("STORAGE_PATH")
    if base == "" {
        base = "./data"
    }
    path := storage.PhotoPath(base, albumID, p.ID, ext)

    // attempt atomic write
    if err := storage.AtomicWrite(path, encodedData); err != nil {
        // ensure DB record is not left behind
        // rollback will remove the inserted row because tx not committed
        return 0, "", fmt.Errorf("atomic write: %w", err)
    }

    // commit transaction now that file exists
    if err := tx.Commit(); err != nil {
        // try to remove file on commit failure
        _ = os.Remove(path)
        return 0, "", fmt.Errorf("commit tx: %w", err)
    }

    return p.ID, path, nil
}
