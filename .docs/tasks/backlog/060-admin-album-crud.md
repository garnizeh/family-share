# Task 060: Admin CRUD — Album Management

**Milestone:** Admin UI  
**Points:** 2 (7 hours)  
**Dependencies:** 020  
**Branch:** `feat/admin-albums`  
**Labels:** `admin`, `crud`

## Description
Implement album CRUD operations: create, list, view, update, and delete albums. Use server-side rendering with HTMX for dynamic updates.

## Acceptance Criteria
- [ ] `GET /admin/albums` — list all albums with cover thumbnails
- [ ] `POST /admin/albums` — create new album
- [ ] `GET /admin/albums/{id}` — view album details and photos
- [ ] `POST /admin/albums/{id}` — update album (title, description, cover)
- [ ] `DELETE /admin/albums/{id}` — delete album and cascade photos
- [ ] HTMX partials for create/update forms and album rows

## Files to Add/Modify
- `internal/handler/admin_albums.go` — album CRUD handlers
- `web/templates/admin/albums_list.html` — album list page
- `web/templates/admin/album_form.html` — create/edit form
- `web/templates/admin/album_row.html` — HTMX partial for single album
- `web/templates/admin/album_detail.html` — album detail view

## Handler Functions
```go
func (h *Handler) ListAlbums(w http.ResponseWriter, r *http.Request)
func (h *Handler) CreateAlbum(w http.ResponseWriter, r *http.Request)
func (h *Handler) ViewAlbum(w http.ResponseWriter, r *http.Request)
func (h *Handler) UpdateAlbum(w http.ResponseWriter, r *http.Request)
func (h *Handler) DeleteAlbum(w http.ResponseWriter, r *http.Request)
```

## Templates
- **albums_list.html**: Grid of albums with title, description, cover, photo count
- **album_form.html**: Modal or inline form (title, description inputs)
- **album_row.html**: Single album card (for HTMX swap on create/update)
- **album_detail.html**: Album header + photo grid

## Tests Required
- [ ] Integration test: create album via POST
- [ ] Integration test: list albums returns created albums
- [ ] Integration test: update album title
- [ ] Integration test: delete album cascades to photos
- [ ] Unit test: form validation (empty title)

## PR Checklist
- [ ] All CRUD operations use sqlc queries
- [ ] HTMX responses return partials, not full pages
- [ ] Delete confirmation implemented (client-side or modal)
- [ ] Empty state shown when no albums exist
- [ ] Tests pass: `go test ./internal/handler/... -v`

## Git Workflow
```bash
git checkout -b feat/admin-albums
# Implement album CRUD handlers and templates
go test ./internal/handler/... -v
# Manual test in browser
git add internal/handler/ web/templates/admin/
git commit -m "feat: implement album CRUD for admin"
git push origin feat/admin-albums
# Open PR: "Add admin album management (CRUD)"
```

## Notes
- Use HTMX `hx-post`, `hx-delete`, `hx-swap` for dynamic updates
- For MVP, no pagination on album list (add later if needed)
- Cover photo selection can use dropdown of existing photos
- Cascade delete handled by SQLite foreign key constraint
