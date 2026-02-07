# Async Upload & Queueing Implementation Plan

## Status: Implementation

The asynchronous upload pipeline and processing queue described below has been implemented in the codebase:

- DB table: `processing_queue` (see `sql/schema/0003_add_processing_queue.sql`).
- Handler: `internal/handler/admin_upload.go` enqueues jobs and returns a progress partial.
- Worker: `internal/worker/worker.go` processes pending jobs, updates job status, and removes temp files.

The remainder of this document describes the original design and the implementation details for maintainers.

## Concept Validation
**Opinion:** Your idea to block new uploads while processing the previous batch is **excellent** and highly recommended for a low-resource VPS.
1.  **Resource Safety:** It acts as a strict "Rate Limiter". By forcing the user to wait, you prevent them from flooding the server with hundreds of concurrent processing jobs that would crash the memory (OOM) or lock the CPU.
2.  **Increased Limits:** Because we are processing sequentially (or with limited concurrency) in the background, we can safely increase the upload limit (e.g., from 5 to **50 files**) because the "heavy lifting" is spread out over time, not hitting the server all at once.

---

## Architecture Overview

1.  **Phase 1: Fast Ingestion**
    *   The `AdminUpload` handler receives the files.
    *   It **only** saves raw files to disk (`tmp_uploads/`) and records a "Job" in the DB.
    *   Complexity: O(N) IO operations (very fast).
    *   Response: Returns a "Progress Bar" UI immediately.

2.  **Phase 2: Background Processing (The Consumer)**
    *   A persistent background worker (or simple goroutine triggered by upload) picks up "Pending" jobs.
    *   It processes images one by one (Process -> Resize -> WebP -> Save to Final DB).
    *   Updates Job status to "Done" or "Error".

3.  **Phase 3: Client Feedback**
    *   The UI polls a status endpoint every few seconds.
    *   Shows "Processing 5/20...".
    *   **Blocks:** The upload form is replaced by this status view.
    *   **Unblocks:** Once all jobs are done, the UI refreshes to show the Album view with the new photos.

---

## Detailed Tasks

### 1. Database Schema
The project now includes a migration that creates `processing_queue`.

If you need to inspect or regenerate the migration, check `sql/schema/0003_add_processing_queue.sql` and the `internal/db` migration runner.
    ```sql
    CREATE TABLE processing_queue (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        album_id INTEGER NOT NULL,
        original_filename TEXT NOT NULL,
        temp_filepath TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'pending', -- pending, processing, completed, failed
        error_message TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE CASCADE
    );
    ```
*   SQLC code was generated and lives under `internal/db/sqlc`.

### 2. Backend Logic (Go)

#### 2.1 Queue Manager (`internal/worker/queue.go`)
*   **Task:** Create a simple worker system.
    *   `Enqueue(albumID, files []File)`: Inserts into DB.
    *   `StartWorker()`: A loop that checks for 'pending' items (or triggered via channel).
    *   **Logic:**
        1.  Fetch next `pending` job.
        2.  Update status to `processing`.
        3.  Run `pipeline.ProcessAndSave`.
        4.  If success: Delete job (or mark `completed`) + Add to `photos` table.
        5.  If fail: Mark `failed` + Save error message.
        6.  Remove temp file.

#### 2.2 Refactor Upload Handler (`admin_upload.go`)
*   **Task:** Rewrite `AdminUploadPhotos`.
    *   **Current:** Loop -> Save Temp -> Process -> Return HTML.
    *   **New:**
        *   Loop -> Save Temp -> **Insert into Processing Queue**.
        *   After loop, trigger Worker (non-blocking).
        *   Return `<div hx-get="/admin/upload/status?album_id=X" ...>` (The Progress View).

#### 2.3 Status Endpoint (`admin_upload_status.go`)
*   **Task:** Create `GET /admin/upload/status`.
    *   Query DB: `SELECT count(*) FROM processing_queue WHERE album_id = ? AND status != 'completed'`.
    *   If `count > 0`: Render progress bar (e.g., "Processing batch... X remaining").
        *   Include `hx-trigger="load delay:1s"` to self-refresh.
    *   If `count == 0`:
        *   Check for errors: `SELECT * FROM processing_queue WHERE status = 'failed'`.
        *   If errors exist: Show error report + "Clear Errors" button.
        *   If no errors: Trigger a redirect or swap to the Album Photos list (`hx-get="/admin/albums/{id}"`).

### 3. Frontend / UI (Templates)

#### 3.0 State Persistence (The "Resume" Feature)
*   **Requirement:** If the admin closes the browser and returns, they must see the progress bar if jobs are still running.
*   **Implementation:**
    *   In the `GetAlbum` handler (`admin_albums.go`), before rendering the page:
    *   Check `SELECT count(*) FROM processing_queue WHERE album_id = ? AND status IN ('pending', 'processing')`.
    *   **If > 0**: Pass a flag (e.g., `ProcessingBatch: true`) to the template.
    *   **Template Logic:**
        ```html
        {{ if .ProcessingBatch }}
            {{ template "upload_progress" . }} <!-- Automatically polling -->
        {{ else }}
            {{ template "upload_form" . }}     <!-- Standard form -->
        {{ end }}
        ```
    *   This ensures the "lock" is persistent and server-side enforced, not just JS state.

#### 3.1 Upload Component (`admin/upload.html`)
*   **Task:** Update the form to handle the "swap".
    *   Form already likely uses `hx-post`.
    *   Ensure `hx-target` points to a container that will be replaced by the progress bar.
    *   **Blocker:** The form disappears, so the user *cannot* upload more until the progress bar goes away.

#### 3.2 Progress Component (`admin/upload_progress.html`)
*   **Task:** Create a new template.
    *   Visual: A simple Tailwind progress bar or spinner.
    *   Text: "Optimizing your photos... please wait."
    *   Logic: Auto-polling via HTMX.

---

## Proposed Roadmap
1.  **Step 1:** Create the DB table (Migration).
2.  **Step 2:** Create the Worker logic (Go).
3.  **Step 3:** Implement the Status Endpoint and Template.
4.  **Step 4:** Switch the Upload Handler to use the Queue.
5.  **Step 5:** Increase upload limits (config).
