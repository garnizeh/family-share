# Task 145: Security — Protect Photo Access

**Milestone:** Security & Privacy  
**Points:** 2 (6 hours)  
**Dependencies:** 080  
**Branch:** `feat/protected-photo-access`  
**Labels:** `security`, `access-control`

## Description
Prevent direct access to `/data/photos/{id}.webp` and ensure photos are only accessible via valid share links or authenticated admin routes.

## Acceptance Criteria
- [ ] Direct access to `/data/photos/{id}.webp` is blocked or requires authorization
- [ ] Public access uses share-token routes (e.g., `/s/{token}/photos/{id}`)
- [ ] Admin can still access photos in the dashboard
- [ ] Access rules enforce token validity and expiry
- [ ] No regression in public share album or photo views

## Files to Add/Modify
- `internal/handler/photo_serve.go` — enforce access checks
- `internal/handler/public_share.go` — add secure photo route
- `internal/handler/routes.go` — update route mappings
- `internal/middleware/auth.go` — use existing auth for admin routes
- `web/templates/public/*` — update image URLs (if needed)

## Implementation Notes
- Add a new handler like `ServeSharedPhoto` that validates share token, target type, and photo ownership.
- Keep admin access under `/admin` protected by session.
- Remove or restrict `/data/photos/{id}.webp` to admin-only or signed URLs.

## Tests Required
- [ ] Integration: unauthenticated access to `/data/photos/{id}.webp` returns 403/404
- [ ] Integration: valid share token can load photo assets
- [ ] Integration: expired/revoked token cannot load photo assets

## PR Checklist
- [ ] Share views still render images correctly
- [ ] Admin UI still renders images correctly
- [ ] No direct static access to raw photo IDs

## Git Workflow
```bash
git checkout -b feat/protected-photo-access
# Implement protected photo routes

go test ./internal/handler/... -v

git add internal/handler/ web/templates/
git commit -m "feat: protect photo access behind share tokens"
git push origin feat/protected-photo-access
```

## Notes
- Consider HTTP 404 for unauthorized access to avoid ID enumeration hints.
