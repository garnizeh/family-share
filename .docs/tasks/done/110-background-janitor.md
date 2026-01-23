# Task 110: Background Janitor — Cleanup Goroutine

**Milestone:** Security & Ops  
**Points:** 2 (6 hours)  
**Dependencies:** 090  
**Branch:** `feat/janitor`  
**Labels:** `ops`, `maintenance`

## Description
Implement a background Goroutine that periodically cleans up expired share links, orphaned photos, expired sessions, and optionally runs SQLite VACUUM.

## Acceptance Criteria
- [ ] Goroutine starts on server startup
- [ ] Runs cleanup tasks every 6 hours (configurable)
- [ ] Deletes expired share links (expires_at < now)
- [ ] Deletes orphaned photos (no album or deleted album)
- [ ] Deletes expired sessions
- [ ] Removes physical files for deleted photos
- [ ] Optional: runs VACUUM monthly to reclaim disk space
- [ ] Graceful shutdown on server stop

## Files to Add/Modify
- `internal/janitor/janitor.go` — cleanup logic and scheduler
- `internal/janitor/janitor_test.go` — unit tests
- `cmd/familyshare/main.go` — start janitor on server init

## Janitor Structure
```go
type Janitor struct {
    db *sql.DB
    storagePath string
    interval time.Duration
    stopChan chan struct{}
}

func (j *Janitor) Start() {
    ticker := time.NewTicker(j.interval)
    for {
        select {
        case <-ticker.C:
            j.runCleanup()
        case <-j.stopChan:
            ticker.Stop()
            return
        }
    }
}

func (j *Janitor) Stop() {
    close(j.stopChan)
}

func (j *Janitor) runCleanup() {
    j.deleteExpiredShareLinks()
    j.deleteOrphanPhotos()
    j.deleteExpiredSessions()
    j.vacuumIfNeeded()
}
```

## Cleanup Tasks
1. **Expired share links**: `DELETE FROM share_links WHERE expires_at < NOW() OR revoked_at IS NOT NULL`
2. **Orphaned photos**: `DELETE FROM photos WHERE album_id NOT IN (SELECT id FROM albums)`
3. **Expired sessions**: `DELETE FROM sessions WHERE expires_at < NOW()`
4. **File cleanup**: For deleted photo rows, remove file from disk
5. **VACUUM**: Run monthly to compact database

## Tests Required
- [ ] Unit test: expired share link deleted
- [ ] Unit test: orphaned photo deleted and file removed
- [ ] Unit test: active share link not deleted
- [ ] Unit test: janitor stops gracefully on signal
- [ ] Integration test: full cleanup cycle

## PR Checklist
- [ ] Cleanup interval configurable via environment
- [ ] Janitor starts as Goroutine, not blocking server
- [ ] Graceful shutdown on SIGTERM/SIGINT
- [ ] File deletion failures logged but don't crash janitor
- [ ] Tests pass: `go test ./internal/janitor/... -v`

## Git Workflow
```bash
git checkout -b feat/janitor
# Implement janitor
go test ./internal/janitor/... -v
git add internal/janitor/ cmd/familyshare/
git commit -m "feat: add background janitor for cleanup tasks"
git push origin feat/janitor
# Open PR: "Implement background janitor for expired data cleanup"
```

## Notes
- Use context for graceful shutdown
- Log cleanup actions (how many rows deleted)
- VACUUM can be slow on large DBs; run during low-traffic hours
- Consider making cleanup tasks pluggable for extensibility
