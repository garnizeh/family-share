# Task 053: Optimize Upload Handler with Disk Streaming

**Milestone:** Admin UI  
**Points:** 3 (12 hours)  
**Dependencies:** 055  
**Branch:** `feat/upload-streaming`  
**Labels:** `optimization`, `performance`, `upload`

## Description
Refactor the admin upload handler to stream uploaded files directly to temporary disk storage instead of buffering them entirely in memory. This prevents Out-Of-Memory issues on low-resource VPS environments when handling large batch uploads.

## Problem Statement
Current implementation (`internal/handler/admin_upload.go`):
- Uses `r.ParseMultipartForm(100MB)` which allocates entire request body in memory
- Reads each file into `bytes.Buffer` before processing
- **Risk:** Multiple concurrent uploads or large batches can exhaust RAM on 512MB-1GB VPS
- **Example:** 10 users uploading 5 photos each (20MB avg) = ~1GB RAM spike

## Solution Architecture
1. **Stream-to-Temp Pattern:**
   - Use `r.MultipartReader()` instead of `ParseMultipartForm`
   - Write each part directly to temp file with `io.Copy(tmpFile, io.LimitReader(part, maxPerFile))`
   - Pass `*os.File` (implements `io.ReadSeeker`) to pipeline
   - Clean up temp file immediately after processing

2. **Temp File Management:**
   - Store in system temp dir (`os.TempDir()`) — cleaned by OS on reboot
   - Use restrictive permissions (`0600`)
   - Remove immediately after `ProcessAndSave` (success or failure)
   - Add background janitor to clean orphaned temp files (for crashes/interruptions)

3. **Resource Limits:**
   - Per-file limit: 25MB (enforced via `io.LimitReader`)
   - Total request limit: 100MB (via `http.MaxBytesReader`)
   - Context timeout: 5 minutes per upload request

## Acceptance Criteria
- [ ] Handler uses `r.MultipartReader()` for streaming parsing
- [ ] Each uploaded file written to temp file in system temp dir (`os.TempDir()`)
- [ ] Temp files have `0600` permissions
- [ ] Temp files cleaned up on success and error paths
- [ ] Per-file size limit (25MB) enforced during copy
- [ ] Total request size limit (100MB) enforced via `MaxBytesReader`
- [ ] Context with timeout applied to prevent hung uploads
- [ ] Background janitor cleans temp files older than 15 minutes
- [ ] Memory usage under 50MB for 10 concurrent 20MB uploads (measured)
- [ ] Existing tests pass without modification

## Files to Add/Modify
- `internal/handler/admin_upload.go` — refactor to streaming
- `internal/storage/cleanup.go` — add temp file janitor (extend existing)
- `internal/handler/admin_upload_test.go` — verify temp cleanup

## Implementation Details

### 1. Refactored Handler (admin_upload.go)
```go
func (h *Handler) AdminUploadPhotos(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
    defer cancel()

    albumIDStr := chi.URLParam(r, "id")
    albumID, err := strconv.ParseInt(albumIDStr, 10, 64)
    if err != nil {
        http.Error(w, "invalid album id", http.StatusBadRequest)
        return
    }

    // Total request size limit
    const maxTotal = int64(100 << 20)
    r.Body = http.MaxBytesReader(w, r.Body, maxTotal)

    // Use MultipartReader for streaming
    mr, err := r.MultipartReader()
    if err != nil {
        http.Error(w, "failed to read multipart", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    flusher, _ := w.(http.Flusher)

    // Use system temp dir (cleaned by OS)
    tmpBaseDir := os.TempDir()

    const maxPerFile = int64(25 << 20) // 25MB

    for {
        part, err := mr.NextPart()
        if err == io.EOF {
            break
        }
        if err != nil {
            // Log and continue
            continue
        }

        // Only process "photos" field
        if part.FormName() != "photos" {
            continue
        }

        var result UploadResult
        result.Filename = part.FileName()

        // Create temp file
        tmp, err := os.CreateTemp(tmpBaseDir, "upload-*.tmp")
        if err != nil {
            result.Error = fmt.Errorf("temp file creation failed")
            h.RenderTemplate(w, "upload_row.html", result)
            if flusher != nil { flusher.Flush() }
            continue
        }
        tmp.Chmod(0600)

        // Copy with size limit
        n, err := io.Copy(tmp, io.LimitReader(part, maxPerFile+1))
        if err != nil {
            tmp.Close()
            os.Remove(tmp.Name())
            result.Error = fmt.Errorf("read failed: %w", err)
            h.RenderTemplate(w, "upload_row.html", result)
            if flusher != nil { flusher.Flush() }
            continue
        }

        if n > maxPerFile {
            tmp.Close()
            os.Remove(tmp.Name())
            result.Error = fmt.Errorf("file too large")
            h.RenderTemplate(w, "upload_row.html", result)
            if flusher != nil { flusher.Flush() }
            continue
        }

        // Seek to beginning for pipeline
        if _, err := tmp.Seek(0, 0); err != nil {
            tmp.Close()
            os.Remove(tmp.Name())
            result.Error = fmt.Errorf("seek failed")
            h.RenderTemplate(w, "upload_row.html", result)
            if flusher != nil { flusher.Flush() }
            continue
        }

        // Process through pipeline
        photo, err := pipeline.ProcessAndSave(ctx, h.db, albumID, tmp, n)
        
        // ALWAYS cleanup temp file
        tmp.Close()
        os.Remove(tmp.Name())

        if err != nil {
            result.Error = err
        } else {
            result.PhotoID = photo.ID
        }

        // Render HTMX partial
        if err := h.RenderTemplate(w, "upload_row.html", result); err != nil {
            http.Error(w, "template render error", http.StatusInternalServerError)
            return
        }
        if flusher != nil {
            flusher.Flush()
        }
    }
}
```

### 2. Background Janitor (storage/cleanup.go extension)
```go
// CleanOrphanedTempFiles removes temp upload files older than maxAge from system temp dir.
// Should be called periodically by a background goroutine.
func CleanOrphanedTempFiles(maxAge time.Duration) error {
    tmpDir := os.TempDir()
    
    entries, err := os.ReadDir(tmpDir)
    if err != nil {
        return err
    }

    cutoff := time.Now().Add(-maxAge)
    
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        
        // Only clean up our upload-*.tmp files (avoid touching other temp files)
        if !strings.HasPrefix(entry.Name(), "upload-") || !strings.HasSuffix(entry.Name(), ".tmp") {
            continue
        }

        fullPath := filepath.Join(tmpDir, entry.Name())
        info, err := entry.Info()
        if err != nil {
            continue
        }

        if info.ModTime().Before(cutoff) {
            os.Remove(fullPath)
        }
    }
    
    return nil
}
```

### 3. Start Janitor in main.go
```go
// In cmd/app/main.go, after server start:
go func() {
    ticker := time.NewTicker(15 * time.Minute)
    defer ticker.Stop()
    for {
        select {
        case <-ticker.C:
            storage.CleanOrphanedTempFiles(15 * time.Minute)
        case <-ctx.Done():
            return
        }
    }
}()
```

## Tests Required
- [ ] Integration test: verify temp file created during upload
- [ ] Integration test: verify temp file removed after successful upload
- [ ] Integration test: verify temp file removed after failed upload (invalid image)
- [ ] Integration test: verify 25MB+ file rejected without filling disk
- [ ] Unit test: `CleanOrphanedTempFiles` removes old files, keeps recent
- [ ] Load test: 10 concurrent uploads, measure peak memory usage

## Performance Benchmarks (Before/After)
Run with `hey` or similar tool:
```bash
# Before (memory buffering)
# Expected: ~1GB RAM for 10x 20MB uploads
# After (disk streaming)  
# Expected: ~50MB RAM for 10x 20MB uploads
```

## PR Checklist
- [ ] All existing handler tests pass
- [ ] New tests for temp file cleanup pass
- [ ] Temp files have restrictive permissions (0600)
- [ ] Janitor goroutine started in main.go
- [ ] Memory usage reduced (measured with pprof or /proc/self/status)
- [ ] No temp file leaks (manual verification after test runs)
- [ ] Error messages user-friendly ("File too large (max 25MB)")
- [ ] Verified `/tmp` is disk-backed (not tmpfs) on target VPS: `df -h /tmp`

## Git Workflow
```bash
git checkout -b feat/upload-streaming
# Implement streaming handler
# Add janitor to storage/cleanup.go
# Update main.go to start janitor
go test ./internal/handler/... -v
go test ./internal/storage/... -v
# Manual test with large file
curl -X POST http://localhost:8080/admin/albums/1/photos \
  -F "photos=@large.jpg" \
  -F "photos=@large2.jpg"
# Verify no temp files remain
ls -la /tmp/upload-*.tmp
git add internal/handler/ internal/storage/ cmd/app/
git commit -m "feat: optimize upload with disk streaming to reduce memory usage"
git push origin feat/upload-streaming
# Open PR: "Optimize upload handler with disk streaming for low-RAM VPS"
```

## Rollback Plan
If streaming causes issues:
- Revert to memory buffering (previous implementation)
- Add memory limit configuration (env var MAX_UPLOAD_MEMORY)
- Consider implementing both modes with feature flag

## Monitoring
After deployment, monitor:
- **Check if `/tmp` is tmpfs:** `df -h /tmp` — if Type=tmpfs, this optimization won't help RAM usage
- Memory usage during peak upload hours (expect <50MB overhead if `/tmp` is disk-backed)
- Disk I/O wait times (streaming adds disk writes)
- Temp directory size (janitor effectiveness): `du -sh /tmp/upload-*.tmp`
- Upload success/failure rates

## Notes
- This optimization is critical for VPS deployments with <2GB RAM
- **Important:** If system `/tmp` is mounted as tmpfs (RAM-backed), this won't reduce RAM usage — verify with `df -h /tmp` before deploying
- On most VPS providers, `/tmp` is disk-backed and cleaned on reboot — ideal for this use case
- Disk streaming trades CPU/memory for I/O — acceptable on SSD VPS
- For extremely high-traffic scenarios, consider object storage (S3) direct uploads
- Temp files survive process crashes — janitor ensures cleanup (but OS also cleans on reboot)
- MultipartReader processes parts sequentially (no parallelism) — safer for low-resource VPS
- Alternative: Make temp dir configurable via env var `TEMP_UPLOAD_DIR` to allow custom paths
