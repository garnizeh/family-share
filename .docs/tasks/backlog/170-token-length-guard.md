# Task 170: Security — Guard Short Tokens

**Milestone:** Security & Ops  
**Points:** 1 (2 hours)  
**Dependencies:** 080  
**Branch:** `feat/token-length-guard`  
**Labels:** `security`, `bug`

## Description
Avoid panics when a token shorter than 8 characters is passed to viewer hash cookie logic.

## Acceptance Criteria
- [ ] No panics on malformed/short token input
- [ ] Cookie name uses safe fallback prefix

## Files to Add/Modify
- `internal/security/viewer_hash.go`

## Implementation Notes
- Guard `token[:8]` with length checks.
- Use a fixed prefix when token is too short.

## Tests Required
- [ ] Unit: short token does not panic and sets cookie name safely

## PR Checklist
- [ ] No new external dependencies

## Git Workflow
```bash
git checkout -b feat/token-length-guard
# Add token length guard

go test ./internal/security/... -v

git add internal/security/

git commit -m "fix: guard short tokens in viewer hash"

git push origin feat/token-length-guard
```
