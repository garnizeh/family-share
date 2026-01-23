package metrics

import (
	"context"
	"database/sql"
	"log"
	"time"

	"familyshare/internal/db/sqlc"
)

// EventType represents the type of activity event
type EventType string

const (
	EventUpload    EventType = "upload"
	EventAlbumView EventType = "album_view"
	EventPhotoView EventType = "photo_view"
	EventShareView EventType = "share_view"
)

// Logger handles activity event logging
type Logger struct {
	queries *sqlc.Queries
}

// New creates a new metrics logger
func New(db *sql.DB) *Logger {
	return &Logger{
		queries: sqlc.New(db),
	}
}

// LogEvent inserts an activity event into the database
func (l *Logger) LogEvent(ctx context.Context, eventType EventType, albumID, photoID, shareLinkID *int64) error {
	var albumIDParam, photoIDParam, shareLinkIDParam sql.NullInt64

	if albumID != nil {
		albumIDParam = sql.NullInt64{Int64: *albumID, Valid: true}
	}
	if photoID != nil {
		photoIDParam = sql.NullInt64{Int64: *photoID, Valid: true}
	}
	if shareLinkID != nil {
		shareLinkIDParam = sql.NullInt64{Int64: *shareLinkID, Valid: true}
	}

	err := l.queries.CreateActivityEvent(ctx, sqlc.CreateActivityEventParams{
		EventType:   string(eventType),
		AlbumID:     albumIDParam,
		PhotoID:     photoIDParam,
		ShareLinkID: shareLinkIDParam,
	})

	if err != nil {
		log.Printf("metrics: failed to log event %s: %v", eventType, err)
	}

	return err
}

// LogUpload logs a photo upload event
func (l *Logger) LogUpload(ctx context.Context, albumID, photoID int64) error {
	return l.LogEvent(ctx, EventUpload, &albumID, &photoID, nil)
}

// LogAlbumView logs an album view event
func (l *Logger) LogAlbumView(ctx context.Context, albumID int64) error {
	return l.LogEvent(ctx, EventAlbumView, &albumID, nil, nil)
}

// LogPhotoView logs a photo view event
func (l *Logger) LogPhotoView(ctx context.Context, photoID int64) error {
	return l.LogEvent(ctx, EventPhotoView, nil, &photoID, nil)
}

// LogShareView logs a share link view event
func (l *Logger) LogShareView(ctx context.Context, shareLinkID int64) error {
	return l.LogEvent(ctx, EventShareView, nil, nil, &shareLinkID)
}

// Stats holds aggregated metrics
type Stats struct {
	Uploads7Days     int64
	Uploads30Days    int64
	AlbumViews7Days  int64
	AlbumViews30Days int64
	PhotoViews7Days  int64
	PhotoViews30Days int64
	ShareViews7Days  int64
	ShareViews30Days int64
}

// GetStats retrieves activity statistics for the dashboard
func (l *Logger) GetStats(ctx context.Context) (*Stats, error) {
	now := time.Now().UTC()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)

	stats := &Stats{}

	// Get 7-day stats
	uploads7, err := l.queries.CountUploadsSince(ctx, sql.NullTime{Time: sevenDaysAgo, Valid: true})
	if err != nil {
		return nil, err
	}
	stats.Uploads7Days = uploads7

	albumViews7, err := l.queries.CountAlbumViewsSince(ctx, sql.NullTime{Time: sevenDaysAgo, Valid: true})
	if err != nil {
		return nil, err
	}
	stats.AlbumViews7Days = albumViews7

	photoViews7, err := l.queries.CountPhotoViewsSince(ctx, sql.NullTime{Time: sevenDaysAgo, Valid: true})
	if err != nil {
		return nil, err
	}
	stats.PhotoViews7Days = photoViews7

	shareViews7, err := l.queries.CountShareViewsSince(ctx, sql.NullTime{Time: sevenDaysAgo, Valid: true})
	if err != nil {
		return nil, err
	}
	stats.ShareViews7Days = shareViews7

	// Get 30-day stats
	uploads30, err := l.queries.CountUploadsSince(ctx, sql.NullTime{Time: thirtyDaysAgo, Valid: true})
	if err != nil {
		return nil, err
	}
	stats.Uploads30Days = uploads30

	albumViews30, err := l.queries.CountAlbumViewsSince(ctx, sql.NullTime{Time: thirtyDaysAgo, Valid: true})
	if err != nil {
		return nil, err
	}
	stats.AlbumViews30Days = albumViews30

	photoViews30, err := l.queries.CountPhotoViewsSince(ctx, sql.NullTime{Time: thirtyDaysAgo, Valid: true})
	if err != nil {
		return nil, err
	}
	stats.PhotoViews30Days = photoViews30

	shareViews30, err := l.queries.CountShareViewsSince(ctx, sql.NullTime{Time: thirtyDaysAgo, Valid: true})
	if err != nil {
		return nil, err
	}
	stats.ShareViews30Days = shareViews30

	return stats, nil
}
