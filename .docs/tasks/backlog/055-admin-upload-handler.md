# Task 055: Admin Upload Handler with HTMX Progress

**Milestone:** Admin UI  
**Points:** 2 (8 hours)  
**Dependencies:** 050  
**Branch:** `feat/admin-upload`  
**Labels:** `admin`, `htmx`, `upload`

## Description
Create the admin upload endpoint that accepts multipart photo uploads, processes them through the pipeline, and returns HTMX partial responses for progress indication.

## Acceptance Criteria
- [ ] `POST /admin/albums/{id}/photos` accepts multipart uploads
- [ ] Supports single and batch uploads
- [ ] Calls image pipeline for each file
- [ ] Returns HTMX partial with success/error status per file
- [ ] Upload size limited to prevent DoS
- [ ] Admin authentication required (middleware)

## Files to Add/Modify
- `internal/handler/admin_upload.go` — upload handler
- `web/templates/admin/upload_row.html` — HTMX response partial
- `internal/middleware/auth.go` — admin auth middleware (stub for now)

## Handler Logic
```go
func (h *Handler) AdminUploadPhotos(w http.ResponseWriter, r *http.Request) {
    albumID := chi.URLParam(r, "id")
    
    // Parse multipart form (max 100MB total)
    r.ParseMultipartForm(100 << 20)
    
    files := r.MultipartForm.File["photos"]
    for _, fileHeader := range files {
        file, _ := fileHeader.Open()
        defer file.Close()
        
        // Process through pipeline
        photo, err := pipeline.ProcessAndSave(ctx, db, albumID, file, maxBytes)
        
        // Return HTMX partial for each file
        tmpl.ExecuteTemplate(w, "upload_row.html", UploadResult{
            Filename: fileHeader.Filename,
            PhotoID: photo.ID,
            Error: err,
        })
    }
}
```

## HTMX Template (upload_row.html)
```html
<div class="upload-row {{ if .Error }}error{{ else }}success{{ end }}">
    <span class="filename">{{ .Filename }}</span>
    {{ if .Error }}
        <span class="status error">Failed: {{ .Error }}</span>
    {{ else }}
        <span class="status success">Uploaded (ID: {{ .PhotoID }})</span>
    {{ end }}
</div>
```

## Tests Required
- [ ] Integration test: upload single JPEG, verify photo created
- [ ] Integration test: upload batch (3 files), verify all created
- [ ] Integration test: upload invalid file, verify error returned
- [ ] Integration test: upload exceeds size limit, verify rejection
- [ ] Unit test: multipart parsing

## PR Checklist
- [ ] Upload size limits enforced (per-file and total)
- [ ] HTMX responses are valid HTML partials
- [ ] Errors are user-friendly ("File too large", not "EOF")
- [ ] Auth middleware applied to route (stub OK for now)
- [ ] Tests pass: `go test ./internal/handler/... -v`
- [ ] Manual test: upload via curl or Postman works

## Git Workflow
```bash
git checkout -b feat/admin-upload
# Implement upload handler
go test ./internal/handler/... -v
# Manual test
git add internal/handler/ web/templates/admin/
git commit -m "feat: add admin photo upload handler with HTMX progress"
git push origin feat/admin-upload
# Open PR: "Implement admin upload handler with HTMX partials"
```

## Notes
- For MVP, admin auth can be a simple stub (always allow)
- Consider streaming response for large batches (flush after each file)
- Use context with timeout to prevent hung uploads
- Log upload attempts and failures for monitoring
