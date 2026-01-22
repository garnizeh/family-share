# Task 080: Public — Share Link Access and View Tracking

**Milestone:** Sharing & View Logic  
**Points:** 2 (8 hours)  
**Dependencies:** 075  
**Branch:** `feat/share-access`  
**Labels:** `public`, `sharing`, `security`

## Description
Implement public share link access with unique visitor tracking, view limits, and expiration enforcement. No login required.

## Acceptance Criteria
- [ ] `GET /s/{token}` — public share link landing page
- [ ] Load share link and validate (not expired, not revoked, views not exceeded)
- [ ] Track unique visitors using signed cookie (viewer_hash)
- [ ] Increment view count only for new visitors
- [ ] Display album or photo based on target_type
- [ ] Show "expired" or "limit reached" pages when applicable

## Files to Add/Modify
- `internal/handler/public_share.go` — share link handler
- `internal/security/viewer_hash.go` — visitor tracking
- `web/templates/public/share_album.html` — album view
- `web/templates/public/share_photo.html` — single photo view
- `web/templates/public/share_expired.html` — error page

## Handler Logic
```go
func (h *Handler) ViewShareLink(w http.ResponseWriter, r *http.Request) {
    token := chi.URLParam(r, "token")
    
    // 1. Load share link
    link := queries.GetShareLinkByToken(token)
    if link == nil || link.RevokedAt != nil {
        renderExpired(w, "Link not found or revoked")
        return
    }
    
    // 2. Check expiration
    if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
        renderExpired(w, "Link has expired")
        return
    }
    
    // 3. Get or create viewer_hash
    viewerHash := getViewerHash(r, token)
    
    // 4. Check view limit
    if link.MaxViews != nil {
        uniqueViews := queries.CountUniqueShareLinkViews(link.ID)
        if uniqueViews >= *link.MaxViews {
            renderExpired(w, "View limit reached")
            return
        }
    }
    
    // 5. Track view (idempotent insert)
    queries.InsertShareLinkView(link.ID, viewerHash)
    
    // 6. Render content
    if link.TargetType == "album" {
        renderAlbum(w, link.TargetID)
    } else {
        renderPhoto(w, link.TargetID)
    }
}
```

## Viewer Hash Logic
- Set signed cookie scoped to token: `_vh_{token_prefix}`
- Cookie contains HMAC(token + IP + User-Agent, secret)
- Cookie is HttpOnly, SameSite=Lax
- Valid for share link lifetime

## Tests Required
- [ ] Integration test: first visit creates viewer_hash and increments view
- [ ] Integration test: second visit from same browser does not increment
- [ ] Integration test: expired link returns 410 Gone
- [ ] Integration test: view limit enforced
- [ ] Integration test: revoked link returns 404

## PR Checklist
- [ ] No authentication required (public access)
- [ ] Viewer tracking prevents refresh abuse
- [ ] Error pages are user-friendly
- [ ] HTMX partials work for photo grids (if album)
- [ ] Tests pass: `go test ./internal/handler/... -v`

## Git Workflow
```bash
git checkout -b feat/share-access
# Implement share link access and tracking
go test ./internal/handler/... -v
git add internal/handler/ internal/security/ web/templates/public/
git commit -m "feat: implement public share link access with view tracking"
git push origin feat/share-access
# Open PR: "Add public share link access and unique view tracking"
```

## Notes
- Use `share_link_views` unique index to prevent duplicate inserts
- Viewer hash prevents simple F5 refresh from consuming views
- For MVP, tracking by cookie is sufficient (no IP fingerprinting)
- Consider rate-limiting share link access (task 085)
