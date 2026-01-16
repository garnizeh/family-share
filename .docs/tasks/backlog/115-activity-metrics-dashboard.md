# Task 115: Activity Metrics Collection and Dashboard

**Milestone:** Metrics & Polish  
**Points:** 2 (7 hours)  
**Dependencies:** 020  
**Branch:** `feat/metrics`  
**Labels:** `admin`, `metrics`

## Description
Implement activity event tracking and a simple admin dashboard showing upload, view, and share statistics for the last 7 and 30 days.

## Acceptance Criteria
- [ ] Activity events logged for: upload, album_view, photo_view, share_view
- [ ] Admin dashboard shows metric cards (last 7/30 days)
- [ ] Metrics: total uploads, album views, photo views, share hits
- [ ] Simple bar chart or table visualization (optional: use Chart.js)
- [ ] Events automatically inserted on relevant actions

## Files to Add/Modify
- `internal/handler/admin_dashboard.go` — dashboard handler
- `internal/metrics/events.go` — event logging helpers
- `web/templates/admin/dashboard.html` — metrics dashboard
- `internal/db/queries/activity_events.sql` — sqlc queries for metrics

## Event Logging
```go
// LogEvent inserts an activity event
func LogEvent(ctx context.Context, db *sql.DB, eventType string, albumID, photoID, shareLinkID *int64) error {
    return queries.InsertActivityEvent(ctx, eventType, albumID, photoID, shareLinkID)
}

// Call from handlers
metrics.LogEvent(ctx, db, "upload", &albumID, &photoID, nil)
metrics.LogEvent(ctx, db, "share_view", nil, nil, &shareLinkID)
```

## Dashboard Queries
```sql
-- name: CountActivityByTypeSince :many
SELECT event_type, COUNT(*) as count
FROM activity_events
WHERE created_at >= ?
GROUP BY event_type;

-- name: CountUploadsSince :one
SELECT COUNT(*) FROM activity_events
WHERE event_type = 'upload' AND created_at >= ?;
```

## Dashboard UI
- Card: "Uploads (Last 7 Days)" — count
- Card: "Uploads (Last 30 Days)" — count
- Card: "Album Views (Last 7 Days)" — count
- Card: "Share Link Hits (Last 30 Days)" — count

## Tests Required
- [ ] Unit test: event insertion
- [ ] Integration test: upload triggers event
- [ ] Integration test: dashboard shows correct counts
- [ ] Unit test: date range filtering (7 vs 30 days)

## PR Checklist
- [ ] Events logged for all relevant actions
- [ ] Dashboard accessible at `/admin/dashboard`
- [ ] Metrics queries optimized (use index on created_at)
- [ ] Dashboard loads quickly (< 100ms)
- [ ] Tests pass: `go test ./internal/... -v`

## Git Workflow
```bash
git checkout -b feat/metrics
# Implement metrics collection and dashboard
go test ./internal/... -v
git add internal/handler/ internal/metrics/ web/templates/admin/ internal/db/queries/
git commit -m "feat: add activity metrics and admin dashboard"
git push origin feat/metrics
# Open PR: "Implement activity tracking and metrics dashboard"
```

## Notes
- For MVP, simple counts are sufficient (no fancy graphs)
- Consider adding "Top Albums" or "Most Shared" (defer to post-MVP)
- Events table can grow large; janitor should clean old events (90+ days)
- Use prepared statements for metrics queries (already handled by sqlc)
