# FamilyShare Project Improvement Report

Date: 2026-01-29 (UTC)

## Scope
I reviewed the Go backend, SSR templates setup, storage pipeline, security, database schema/queries, middleware, and operational configuration. The focus was on storage efficiency, security, reliability, UX, and maintainability.

## Key Strengths (Keep)
- Clear MVP scope and minimal stack (Go + SQLite + SSR + HTMX).
- Image pipeline enforces resize + WebP conversion and avoids originals.
- Rate limiting on share links and admin endpoints.
- Janitor background cleanup for expired links, sessions, old activity events.
- Strong use of UTC across most time handling.

## Critical Improvements (High Risk / Must Fix)
1. **Public photo URL is guessable and bypasses share-link access control**
   - `GET /data/photos/{id}.webp` is publicly accessible and does not verify share token or permissions. Anyone can enumerate IDs.
   - **Fix**: Serve photos only through share routes (e.g., `/s/{token}/photos/{id}`) or use signed/opaque URLs. Block direct access to `/data/photos/*` or validate access before serving.

2. **Photo storage path changes over time**
   - `storage.PhotoPath()` uses `time.Now().UTC()` for path layout. `ServePhoto()` also uses current time. Once the month changes, previously stored photos become unreachable.
   - **Fix**: Store the storage path or the `created_at` year/month in DB and build path from that. Alternatively store the full path in the `photos` table.

3. **Inconsistent storage base path**
   - `SaveProcessedImage()` reads `STORAGE_PATH` from env directly, while handlers use `storage.New(cfg.DataDir)` and `storage.PhotoPath(h.storage.BaseDir, ...)`.
   - **Fix**: Pass the configured base dir into the pipeline or read from `cfg.DataDir` consistently.

## Security Improvements (High / Medium)
4. **Hard-coded viewer hash secret**
   - `ViewerHashSecret` is compiled into the binary.
   - **Fix**: Load from environment and fail fast if missing in production. Document it in `.env.example`.

5. **Viewer hash cookie security flags**
   - `SetViewerHashCookie()` sets `Secure: false` unconditionally.
   - **Fix**: set `Secure` based on config (`FORCE_HTTPS`), not `r.TLS` (proxies terminate TLS). Consider `SameSite=Strict` for better protection.

6. **Token cookie name assumes token length >= 8**
   - `token[:8]` can panic if token is malformed or short.
   - **Fix**: guard length and fallback to a safe prefix.

7. **Admin CSRF protection missing**
   - Admin POST/PUT/DELETE are not CSRF-protected.
   - **Fix**: add CSRF tokens to admin forms (HTMX-friendly) or use double-submit cookies.

## Reliability & Data Integrity
8. **Use of request context in goroutine**
   - `ViewShareLink` logs metrics with `r.Context()` inside a goroutine; it may be canceled early.
   - **Fix**: use `context.Background()` or a short timeout context in goroutines.

9. **Temporary file cleanup ignores `TEMP_UPLOAD_DIR`**
   - Janitor uses `os.TempDir()` only, but uploads can be stored in `TEMP_UPLOAD_DIR`.
   - **Fix**: pass the temp dir to `CleanOrphanedTempFiles()` or add config to janitor.

10. **Rate limiter IP trust model**
    - Uses `X-Forwarded-For` without validating trusted proxy ranges.
    - **Fix**: allow a configured list of trusted proxies and only honor forwarded headers when coming from them.

## Storage & Performance
11. **Database schema allows `avif` but pipeline produces only WebP**
    - `photos.format` check includes `avif` but no AVIF pipeline support.
    - **Fix**: either implement AVIF encoding or remove `avif` from schema constraints.

12. **Static file caching**
    - Static assets and photos are served without cache headers.
    - **Fix**: add `Cache-Control` headers (immutable for versioned assets) to improve client performance.

## UX / Admin
13. **Error messaging on upload pipeline**
    - Upload errors are generic; users may not understand failures.
    - **Fix**: provide human-friendly error strings (e.g., “Unsupported format”, “Image too large”, “Decode failed”).

14. **Health check does not verify DB**
    - `/health` is a simple endpoint (not shown here), but DB connectivity is not validated.
    - **Fix**: include a lightweight DB ping to detect storage failures.

## Configuration & Ops
15. **`.env` lacks some config keys**
    - `STORAGE_PATH`, `TEMP_UPLOAD_DIR`, and `JANITOR_INTERVAL` are missing in `.env`. `VIEWER_HASH_SECRET` not documented.
    - **Fix**: align `.env` and `.env.example` with all supported env vars.

16. **Logging verbosity**
    - Template loading logs every template name on startup. Fine for dev but noisy in production.
    - **Fix**: guard with a debug flag.

## Tests & Tooling
17. **Add tests for security-critical paths**
    - Suggested tests:
      - Attempt to access `/data/photos/{id}.webp` without share token.
      - Ensure correct storage path resolution across month boundaries.
      - Validate viewer hash for short tokens.
      - Verify rate-limiter honors trusted proxy configuration.

## Quick Wins (Low Effort / High Value)
- Move viewer hash secret to env and document it.
- Ensure photo path derivation is stable (store path or created_at month/year).
- Block direct photo serving or require share token.
- Use config-based `Secure` cookies when behind proxy.

## Long-Term Enhancements
- Add optional AVIF output with feature flag.
- Add lightweight background job metrics and reporting.
- Add optional album-level download packages (ZIP) with rate limit and max size.

---

If you want, I can turn these into a prioritized task list or open PR-ready changes.
