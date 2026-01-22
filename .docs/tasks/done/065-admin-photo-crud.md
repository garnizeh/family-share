# Task 065: Admin CRUD — Photo Management

**Milestone:** Admin UI  
**Points:** 1 (5 hours)  
**Dependencies:** 060  
**Branch:** `feat/admin-photos`  
**Labels:** `admin`, `crud`

## Description
Implement photo management: list photos in album, delete individual photos, and optionally set album cover photo.

## Acceptance Criteria
- [ ] `GET /admin/albums/{id}/photos` — list photos in album (paginated)
- [ ] `DELETE /admin/photos/{id}` — delete photo (file + DB record)
- [ ] `POST /admin/albums/{id}/cover` — set cover photo
- [ ] HTMX partials for photo grid and deletion
- [ ] Physical file deleted when photo record removed

## Files to Add/Modify
- `internal/handler/admin_photos.go` — photo management handlers
- `web/templates/admin/photo_grid.html` — photo grid partial
- `web/templates/admin/photo_card.html` — single photo card with actions
- `internal/storage/delete.go` — file deletion helper

## Handler Functions
```go
func (h *Handler) ListAlbumPhotos(w http.ResponseWriter, r *http.Request)
func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request)
func (h *Handler) SetAlbumCover(w http.ResponseWriter, r *http.Request)
```

## Deletion Logic
```go
func DeletePhoto(ctx context.Context, db *sql.DB, photoID int64) error {
    // 1. Get photo record (to get filename)
    photo := queries.GetPhoto(photoID)
    
    // 2. Delete DB record
    queries.DeletePhoto(photoID)
    
    // 3. Delete physical file
    os.Remove(photo.Filename)
    
    return nil
}
```

## Tests Required
- [ ] Integration test: delete photo removes file and DB record
- [ ] Integration test: set cover photo updates album
- [ ] Integration test: delete photo that is album cover (cover set to null)
- [ ] Unit test: file deletion error handling

## PR Checklist
- [ ] Physical files deleted when photos removed
- [ ] Delete operation is idempotent (safe to retry)
- [ ] HTMX response removes photo card from grid
- [ ] Cover photo selection updates album immediately
- [ ] Tests pass: `go test ./internal/handler/... -v`

## Git Workflow
```bash
git checkout -b feat/admin-photos
# Implement photo management
go test ./internal/handler/... -v
git add internal/handler/ internal/storage/ web/templates/admin/
git commit -m "feat: implement photo management for admin"
git push origin feat/admin-photos
# Open PR: "Add admin photo management (delete, set cover)"
```

## Notes
- For MVP, no photo editing (crop, rotate) — accept as-is
- Log file deletion failures but don't block DB deletion
