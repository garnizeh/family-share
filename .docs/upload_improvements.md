# Upload & Processing Improvements

This document outlines the diagnosis and proposed solutions for `SQLITE_BUSY` errors and connection timeouts during photo uploads.

## 1. SQLite Busy Errors (`SQLITE_BUSY`)

### Diagnosis
The logs showed `create photo record: database is locked (5) (SQLITE_BUSY)` in earlier runs. This happens because SQLite's default locking can cause write contention if long-running writes overlap with other writers.

### Resolution implemented in code
The application now enables WAL mode and sets a busy timeout during DB initialization to reduce transient SQLITE_BUSY failures.

See `internal/db/db.go` for the implemented PRAGMA changes (WAL + busy_timeout).

### 2. Upload Timeouts & Asynchronous Processing

### Diagnosis
Previously the upload handler was fully synchronous (read -> process -> respond) which sometimes caused connection timeouts for slow or large uploads.
1.  Read file part from request.
2.  Save to temp disk.
3.  **Process image (heavy CPU/IO operation).**
4.  Save to DB.
5.  Send HTML response.
6.  *Repeat for next file.*

If a user uploads 5 photos, and each takes 20 seconds to process, the browser/connection waits 100 seconds to send all data, likely triggering the `i/o timeout` seen in the logs (`read failed: read tcp ... i/o timeout`).

### Strategy: Asynchronous Processing Queue (Implemented)
The codebase now decouples ingestion from processing. `AdminUploadPhotos` saves incoming files to a temp directory and inserts jobs into a `processing_queue` table. A background worker processes jobs and updates status.

#### Step 1: Client-Side Upload Limit
To immediately mitigate memory/timeout pressure, limit the `input` to allow max 5 files, or use a JavaScript snippet to chunk uploads.
*   **HTML:** `<input type="file" name="photos" multiple accept="..." onchange="if(this.files.length > 5) { alert('Max 5 files'); this.value=''; }">`
*   **Better:** Use HTMX `hx-encoding="multipart/form-data"` which sends files in one request, but we should handle them faster.

#### Step 2: Implementation Plan (Async)

1.  **Current implementation:**
    * `internal/handler/admin_upload.go` enqueues jobs after saving temp files.
    * `internal/worker/worker.go` consumes the queue and runs the image pipeline.
    * The user-facing UI shows a polling progress partial until processing completes.

2.  **UI Updates (HTMX):**
    *   The "Processing..." card needs to poll for completion.
    *   Return HTML: `<div hx-get="/admin/photos/status/{temp_id}" hx-trigger="load delay:2s" hx-swap="outerHTML">Processing...</div>`
    *   When the background job finishes, the polling endpoint returns the final `upload_row.html` (the thumbnail).

### Proposed Immediate "Quick Fix" (Batch of 5)
If fully async is too complex for now, we can optimize the existing loop to be safer:
1.  **Phase A (Read):** Read *all* 5 multipart files to temp disk first. This drains the request body quickly, preventing the `read tcp` timeout.
2.  **Phase B (Process):** Loop through the saved temp files and process them one by one (or in parallel), then send the response.

*Note: Phase B still delays the HTTP response, so the browser might timeout waiting for the *response* header. Async/Polling is preferred.*

### Summary
- WAL and busy timeout were applied to reduce SQLITE_BUSY errors.
- Upload handler and processing queue were implemented to make uploads robust for low-resource servers.
- For operational issues (timeouts), ensure your reverse proxy timeouts are aligned with server settings (see `.docs/deployment/reverse-proxy.md`).
