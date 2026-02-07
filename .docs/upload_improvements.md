# Upload & Processing Improvements

This document outlines the diagnosis and proposed solutions for `SQLITE_BUSY` errors and connection timeouts during photo uploads.

## 1. SQLite Busy Errors (`SQLITE_BUSY`)

### Diagnosis
The logs show `create photo record: database is locked (5) (SQLITE_BUSY)`. This happens because the default SQLite mode allows only one writer at a time, and without a "busy timeout," any concurrent write attempt immediately fails if the lock is held.

### Solution: Enable WAL Mode & Busy Timeout
We need to configure SQLite to use **Write-Ahead Logging (WAL)**, which allows concurrent readers, and set a **busy timeout** so the application waits for the lock instead of failing immediately.

**Recommended Changes in `internal/db/db.go`:**

```go
func InitDB(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", path)
    // ...
    
    // Enable WAL mode (better concurrency)
    if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
        db.Close()
        return nil, fmt.Errorf("enable wal: %w", err)
    }

    // specific busy timeout (e.g., 5000ms = 5s)
    if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
        db.Close()
        return nil, fmt.Errorf("set busy timeout: %w", err)
    }

    // ... existing foreign_keys logic ...
}
```

## 2. Upload Timeouts & Synchronous Processing

### Diagnosis
The current upload handler (`AdminUploadPhotos`) is fully **synchronous**:
1.  Read file part from request.
2.  Save to temp disk.
3.  **Process image (heavy CPU/IO operation).**
4.  Save to DB.
5.  Send HTML response.
6.  *Repeat for next file.*

If a user uploads 5 photos, and each takes 20 seconds to process, the browser/connection waits 100 seconds to send all data, likely triggering the `i/o timeout` seen in the logs (`read failed: read tcp ... i/o timeout`).

### Strategy: Asynchronous Processing Queue
To fix this robustly, we must decouple the *upload* (receiving bytes) from the *processing* (resize/encode).

#### Step 1: Client-Side Upload Limit
To immediately mitigate memory/timeout pressure, limit the `input` to allow max 5 files, or use a JavaScript snippet to chunk uploads.
*   **HTML:** `<input type="file" name="photos" multiple accept="..." onchange="if(this.files.length > 5) { alert('Max 5 files'); this.value=''; }">`
*   **Better:** Use HTMX `hx-encoding="multipart/form-data"` which sends files in one request, but we should handle them faster.

#### Step 2: Implementation Plan (Async)

1.  **Modify `admin_upload.go`:**
    *   Loop through multipart files.
    *   **Only** save them to `tmp_uploads/` and validation (is it an image?).
    *   Do **not** call `pipeline.ProcessAndSaveWithFormat` immediately.
    *   Instead, start a **Goroutine** for each file (or feed a worker channel) to process the file in the background.
    *   Return a "Processing..." UI card immediately to the user.

2.  **UI Updates (HTMX):**
    *   The "Processing..." card needs to poll for completion.
    *   Return HTML: `<div hx-get="/admin/photos/status/{temp_id}" hx-trigger="load delay:2s" hx-swap="outerHTML">Processing...</div>`
    *   When the background job finishes, the polling endpoint returns the final `upload_row.html` (the thumbnail).

### Proposed Immediate "Quick Fix" (Batch of 5)
If fully async is too complex for now, we can optimize the existing loop to be safer:
1.  **Phase A (Read):** Read *all* 5 multipart files to temp disk first. This drains the request body quickly, preventing the `read tcp` timeout.
2.  **Phase B (Process):** Loop through the saved temp files and process them one by one (or in parallel), then send the response.

*Note: Phase B still delays the HTTP response, so the browser might timeout waiting for the *response* header. Async/Polling is preferred.*

### Summary of Tasks
1.  [ ] **DB:** Update `InitDB` with WAL & Busy Timeout.
2.  [ ] **Handler:** Refactor `AdminUploadPhotos` to:
    *   Read/Save to temp disk immediately (drain request).
    *   Launch processing in background (goroutine).
    *   Respond with a "pending" template that polls for completion.
3.  [ ] **Frontend:** Add polling endpoint (`/admin/upload/status/{id}`) to check if the photo is ready.
