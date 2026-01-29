# Task 165: Security — Secure Viewer Hash Cookie

**Milestone:** Security & Ops  
**Points:** 1 (3 hours)  
**Dependencies:** 160  
**Branch:** `feat/secure-viewer-cookie`  
**Labels:** `security`, `cookies`

## Description
Set `Secure` and `SameSite` flags based on configuration rather than assuming direct TLS termination.

## Acceptance Criteria
- [ ] Cookie `Secure` flag is configurable (e.g., `FORCE_HTTPS`)
- [ ] `SameSite` is adjustable (default Lax)
- [ ] Works correctly behind reverse proxy

## Files to Add/Modify
- `internal/security/viewer_hash.go`
- `internal/config/config.go`
- `.env.example`

## Implementation Notes
- Add config: `ForceHTTPS` and `CookieSameSite` (enum or string).
- Use those values when setting cookies for viewer hashes and sessions.

## Tests Required
- [ ] Unit: cookie flags are set as configured

## PR Checklist
- [ ] Defaults safe for production
- [ ] No breaking changes for local dev

## Git Workflow
```bash
git checkout -b feat/secure-viewer-cookie
# Implement configurable cookie flags

go test ./internal/security/... -v

git add internal/ .env.example

git commit -m "feat: make cookie security flags configurable"

git push origin feat/secure-viewer-cookie
```
