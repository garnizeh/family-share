# Task 160: Security — Viewer Hash Secret from Env

**Milestone:** Security & Ops  
**Points:** 1 (3 hours)  
**Dependencies:** 080  
**Branch:** `feat/viewer-hash-secret`  
**Labels:** `security`, `config`

## Description
Load the viewer hash HMAC secret from environment/config instead of hard-coding it.

## Acceptance Criteria
- [ ] Viewer hash secret is configurable via env
- [ ] Missing secret logs a warning in dev; hard fail in prod (configurable)
- [ ] `.env.example` documents the variable

## Files to Add/Modify
- `internal/security/viewer_hash.go` — use config-based secret
- `internal/config/config.go` — add `ViewerHashSecret`
- `.env.example` — add `VIEWER_HASH_SECRET`

## Tests Required
- [ ] Unit: viewer hash changes when secret changes
- [ ] Unit: empty secret handled according to environment

## PR Checklist
- [ ] No hard-coded secrets in code
- [ ] Docs updated

## Git Workflow
```bash
git checkout -b feat/viewer-hash-secret
# Load viewer hash secret from env

go test ./internal/security/... -v

git add internal/ .env.example

git commit -m "feat: load viewer hash secret from env"

git push origin feat/viewer-hash-secret
```
