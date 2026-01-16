# Task 100: Public UX — Album Gallery with HTMX Pagination

**Milestone:** Admin UI  
**Points:** 2 (7 hours)  
**Dependencies:** 080  
**Branch:** `feat/public-gallery`  
**Labels:** `public`, `htmx`, `ux`

## Description
Build the public album gallery view with HTMX-powered infinite scroll or click-to-load pagination for smooth photo browsing.

## Acceptance Criteria
- [ ] Album page renders initial grid of photos (e.g., 20 photos)
- [ ] "Load More" button or scroll-trigger fetches next page via HTMX
- [ ] HTMX partial appends new photos to grid
- [ ] Responsive grid layout (1-4 columns based on screen size)
- [ ] Lazy-load images for performance
- [ ] Empty state shown for albums with no photos

## Files to Add/Modify
- `internal/handler/public_gallery.go` — gallery handler with pagination
- `web/templates/public/album_gallery.html` — main gallery page
- `web/templates/public/photo_grid_partial.html` — HTMX partial for photos
- `web/static/styles.css` — Tailwind or custom grid styles

## Handler Logic
```go
func (h *Handler) AlbumGallery(w http.ResponseWriter, r *http.Request) {
    albumID := chi.URLParam(r, "id")
    page := r.URL.Query().Get("page")
    pageNum, _ := strconv.Atoi(page)
    if pageNum < 1 {
        pageNum = 1
    }
    
    limit := 20
    offset := (pageNum - 1) * limit
    
    photos := queries.ListAlbumPhotos(albumID, limit, offset)
    hasMore := len(photos) == limit
    
    if isHTMXRequest(r) {
        renderPartial(w, "photo_grid_partial.html", photos, hasMore, pageNum+1)
    } else {
        renderFull(w, "album_gallery.html", photos, hasMore, pageNum+1)
    }
}
```

## HTMX Markup
```html
<!-- Load More Button -->
<button hx-get="/a/{{ .AlbumID }}/photos?page={{ .NextPage }}" 
        hx-swap="beforeend" 
        hx-target="#photo-grid">
    Load More
</button>

<!-- Infinite Scroll Trigger (alternative) -->
<div hx-get="/a/{{ .AlbumID }}/photos?page={{ .NextPage }}" 
     hx-trigger="revealed" 
     hx-swap="afterend">
</div>
```

## Tests Required
- [ ] Integration test: first page loads initial photos
- [ ] Integration test: HTMX request loads next page and appends
- [ ] Integration test: no "Load More" shown when all photos loaded
- [ ] Unit test: pagination math (offset, limit)
- [ ] Manual test: infinite scroll works smoothly

## PR Checklist
- [ ] Photo grid is responsive (mobile, tablet, desktop)
- [ ] Images lazy-load (loading="lazy" attribute)
- [ ] HTMX partials append without duplicates
- [ ] Empty state shown for empty albums
- [ ] Tests pass: `go test ./internal/handler/... -v`

## Git Workflow
```bash
git checkout -b feat/public-gallery
# Implement gallery with HTMX pagination
go test ./internal/handler/... -v
# Manual test in browser
git add internal/handler/ web/templates/public/ web/static/
git commit -m "feat: add public album gallery with HTMX pagination"
git push origin feat/public-gallery
# Open PR: "Implement album gallery with HTMX infinite scroll"
```

## Notes
- Use Tailwind grid classes for responsive layout
- Lazy-loading improves initial page load time
- Consider adding photo count indicator ("Showing 20 of 156")
- Defer lightbox/carousel to task 105
