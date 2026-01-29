# Task 175: Ops — Janitor Cleans TEMP_UPLOAD_DIR

**Milestone:** Ops & Maintenance  
**Points:** 1 (3 hours)  
**Dependencies:** 110  
**Branch:** `feat/janitor-temp-dir`  
**Labels:** `ops`, `cleanup`

## Description
Ensure janitor cleanup uses the configured `TEMP_UPLOAD_DIR`, not only `os.TempDir()`.

## Acceptance Criteria
- [ ] Temp cleanup checks the configured temp upload directory
- [ ] Default remains OS temp dir if not set

## Files to Add/Modify
- `internal/janitor/janitor.go`
- `internal/storage/cleanup.go`
- `internal/config/config.go`

## Implementation Notes
- Accept a temp dir in janitor config.
- Update cleanup helper to accept a directory path.

## Tests Required
- [ ] Unit: cleanup removes old temp files in custom temp dir

## PR Checklist
- [ ] Config documented in `.env.example`

## Git Workflow
```bash
git checkout -b feat/janitor-temp-dir
# Implement temp dir cleanup

go test ./internal/janitor/... -v

git add internal/

git commit -m "feat: janitor cleans configured temp upload dir"

git push origin feat/janitor-temp-dir
```
