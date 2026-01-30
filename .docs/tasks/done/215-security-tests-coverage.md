# Task 215: Testing — Security-Critical Paths

**Milestone:** Testing & QA  
**Points:** 2 (6 hours)  
**Dependencies:** 120  
**Branch:** `test/security-critical-paths`  
**Labels:** `tests`, `security`

## Description
Add tests that cover security-critical paths identified in the improvement report.

## Acceptance Criteria
- [ ] Direct photo access without token is denied
- [ ] Share links enforce expiration and max views
- [ ] Viewer hash handles short tokens safely
- [ ] Rate limiter honors trusted proxies config

## Files to Add/Modify
- `internal/handler/public_share_test.go`
- `internal/handler/photo_serve_test.go`
- `internal/security/viewer_hash_test.go`
- `internal/middleware/ratelimit_test.go`

## Tests Required
- [ ] Integration: unauthenticated `/data/photos/{id}.webp` blocked
- [ ] Integration: expired share link blocked
- [ ] Unit: short token does not panic
- [ ] Unit: trusted proxy header handling

## PR Checklist
- [ ] Tests are deterministic and fast
- [ ] No external services required

## Git Workflow
```bash
git checkout -b test/security-critical-paths
# Add security-critical tests

go test ./internal/... -v

git add internal/

git commit -m "test: add security-critical path coverage"

git push origin test/security-critical-paths
```
