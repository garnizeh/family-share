# Task 200: Security — Admin CSRF Protection

**Milestone:** Security & Ops  
**Points:** 2 (6 hours)  
**Dependencies:** 090  
**Branch:** `feat/admin-csrf`  
**Labels:** `security`, `csrf`

## Description
Add CSRF protection for all admin state-changing requests (POST/PUT/DELETE), including HTMX requests.

## Acceptance Criteria
- [ ] CSRF tokens generated per session
- [ ] Admin forms include CSRF token (hidden input)
- [ ] HTMX requests include CSRF header automatically
- [ ] Invalid or missing CSRF token returns 403
- [ ] Does not block GET/HEAD requests

## Files to Add/Modify
- `internal/middleware/csrf.go`
- `internal/handler/routes.go`
- `web/templates/layout/*` (CSRF meta)
- `web/templates/admin/*` (hidden input fields)

## Implementation Notes
- Token can be signed HMAC stored in cookie or session.
- Add small JS snippet to set `X-CSRF-Token` for HTMX.

## Tests Required
- [ ] Unit: valid CSRF passes
- [ ] Unit: invalid CSRF fails
- [ ] Integration: POST without CSRF returns 403
- [ ] Integration: HTMX request with header passes

## PR Checklist
- [ ] Works with HTMX forms
- [ ] Documented in `.docs/`

## Git Workflow
```bash
git checkout -b feat/admin-csrf
# Implement CSRF protection

go test ./internal/middleware/... -v

go test ./internal/handler/... -v

git add internal/ web/templates/

git commit -m "feat: add CSRF protection for admin routes"

git push origin feat/admin-csrf
```

## Notes
- Keep implementation simple for MVP but secure.
