# Task 075: Admin — Create and Manage Share Links

**Milestone:** Sharing & View Logic  
**Points:** 2 (7 hours)  
**Dependencies:** 070, 060  
**Branch:** `feat/admin-shares`  
**Labels:** `admin`, `sharing`

## Description
Allow admin to create, list, and revoke share links for albums and photos. Include UI for setting max views and expiration time.

## Acceptance Criteria
- [ ] `POST /admin/shares` — create share link with options
- [ ] `GET /admin/shares` — list all active share links
- [ ] `DELETE /admin/shares/{id}` — revoke share link
- [ ] Form to select target (album/photo), max views, expiration
- [ ] Share link URL displayed for copying
- [ ] Revoked links clearly marked

## Files to Add/Modify
- `internal/handler/admin_shares.go` — share link management
- `web/templates/admin/shares_list.html` — list of share links
- `web/templates/admin/share_form.html` — create form
- `web/templates/admin/share_row.html` — HTMX partial for single share

## Handler Functions
```go
func (h *Handler) CreateShareLink(w http.ResponseWriter, r *http.Request)
func (h *Handler) ListShareLinks(w http.ResponseWriter, r *http.Request)
func (h *Handler) RevokeShareLink(w http.ResponseWriter, r *http.Request)
```

## Create Form Fields
- **Target Type**: radio (Album / Photo)
- **Target ID**: dropdown (select album or photo)
- **Max Views**: number input (optional, blank = unlimited)
- **Expires At**: datetime-local input (optional, blank = never)

## Share Link URL Format
```
https://yoursite.com/s/{token}
```

## Tests Required
- [ ] Integration test: create share link, verify token generated
- [ ] Integration test: list share links shows created link
- [ ] Integration test: revoke share link sets revoked_at
- [ ] Unit test: form validation (invalid target_id)
- [ ] Integration test: expired link not shown in active list

## PR Checklist
- [ ] Share link URLs are copyable (click-to-copy JS or readonly input)
- [ ] HTMX updates list on create/revoke
- [ ] Form validation prevents invalid inputs
- [ ] Tests pass: `go test ./internal/handler/... -v`
- [ ] UI clearly shows expiration and view limit

## Git Workflow
```bash
git checkout -b feat/admin-shares
# Implement share link management
go test ./internal/handler/... -v
git add internal/handler/ web/templates/admin/
git commit -m "feat: add admin share link management"
git push origin feat/admin-shares
# Open PR: "Implement admin share link creation and management"
```

## Notes
- Use `GenerateSecureToken()` from task 070
- For MVP, admin creates share links manually (no bulk or auto-share)
- Copy-to-clipboard can use simple JS snippet or Alpine.js
- Consider showing QR code for each link (defer to post-MVP)
