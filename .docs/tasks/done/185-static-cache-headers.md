# Task 185: Performance — Static Cache Headers

**Milestone:** Performance & UX  
**Points:** 1 (3 hours)  
**Dependencies:** 052  
**Branch:** `feat/static-cache-headers`  
**Labels:** `performance`, `http`

## Description
Add `Cache-Control` headers for static assets and photo responses to improve client performance and reduce bandwidth.

## Acceptance Criteria
- [ ] Static assets served with long-lived cache headers
- [ ] Photo responses include appropriate caching headers
- [ ] No cache headers for dynamic HTML pages

## Files to Add/Modify
- `internal/handler/routes.go`
- `internal/handler/photo_serve.go`

## Implementation Notes
- Use middleware or wrapper to set `Cache-Control` on `/static/*` and photo routes.
- For immutable assets, consider `Cache-Control: public, max-age=31536000, immutable`.

## Tests Required
- [ ] Unit: response headers include cache-control for static files

## PR Checklist
- [ ] No negative impact on admin pages

## Git Workflow
```bash
git checkout -b feat/static-cache-headers
# Add cache headers for static assets and photos

go test ./internal/handler/... -v

git add internal/handler/

git commit -m "perf: add cache headers for static assets"

git push origin feat/static-cache-headers
```
