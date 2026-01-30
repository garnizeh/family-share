# Task 205: UX — Friendly Upload Error Messages

**Milestone:** UX & Admin  
**Points:** 1 (3 hours)  
**Dependencies:** 055  
**Branch:** `feat/upload-error-messages`  
**Labels:** `ux`, `upload`

## Description
Improve user-facing error messages during photo uploads to be clear and actionable.

## Acceptance Criteria
- [ ] Common failures mapped to friendly messages
- [ ] Errors displayed in upload results (HTMX row)
- [ ] Technical details kept in logs only

## Files to Add/Modify
- `internal/handler/admin_upload.go`
- `internal/pipeline/*` (map/return typed errors)
- `web/templates/admin/*` (upload row styling)

## Implementation Notes
- Map errors like: unsupported format, too large, decode failed.
- Keep existing error logs for debugging.

## Tests Required
- [ ] Unit: error mapping returns friendly message
- [ ] Integration: invalid file shows user-friendly error

## PR Checklist
- [ ] No change to pipeline behavior, only messaging

## Git Workflow
```bash
git checkout -b feat/upload-error-messages
# Improve upload error messages

go test ./internal/handler/... -v

git add internal/ web/templates/

git commit -m "feat: improve upload error messaging"

git push origin feat/upload-error-messages
```
