# Task 150: Storage — Stable Photo Paths

**Milestone:** Storage & Data Integrity  
**Points:** 2 (6 hours)  
**Dependencies:** 050  
**Branch:** `feat/stable-photo-paths`  
**Labels:** `storage`, `db`

## Description
Ensure stored photos remain accessible across month/year boundaries by deriving storage paths from persisted data instead of `time.Now()`.

## Acceptance Criteria
- [ ] Photo paths are stable and independent of current time
- [ ] Stored photos are accessible after month/year changes
- [ ] Migration approach documented if DB changes are required

## Files to Add/Modify
- `internal/storage/paths.go` — path generation
- `internal/pipeline/save.go` — store derived path or date parts
- `internal/handler/photo_serve.go` — resolve correct storage path
- `sql/schema/0003_add_photo_path.sql` (if storing path in DB)
- `sql/queries/photos.sql` — queries updated if schema changes

## Implementation Options
1. **Store full path** in `photos.storage_path` on creation.
2. **Store year/month** in `photos.created_at` and derive path from timestamp.

## Tests Required
- [ ] Unit: path derivation matches stored values
- [ ] Integration: photo accessible after simulated month boundary

## PR Checklist
- [ ] Backfill existing rows if new column is introduced
- [ ] No breaking changes to existing routes

## Git Workflow
```bash
git checkout -b feat/stable-photo-paths
# Implement stable path storage

go test ./internal/storage/... -v

go test ./internal/handler/... -v

git add internal/ sql/
git commit -m "feat: make photo paths stable across time"
git push origin feat/stable-photo-paths
```

## Notes
- Prefer storing a path or date parts to avoid implicit time dependencies.
