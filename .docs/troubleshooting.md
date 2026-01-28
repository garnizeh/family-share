# Troubleshooting

## Login fails
- Ensure `ADMIN_PASSWORD_HASH` is set to a bcrypt hash (not plain text).
- Regenerate hash using `make hash-password`.

## Photos not uploading
- Verify `STORAGE_PATH`/`DATA_DIR` are writable.
- Check disk space on the VPS.
- If using Docker, ensure volumes are mounted correctly.

## Share link shows expired
- Check `expires_at` and `max_views` limits.
- Create a new share link if needed.

## High disk usage
- Clean old data under `data/photos/` if safe.
- Reduce upload sizes or run janitor more frequently.
- Consider `VACUUM` on the SQLite database during maintenance windows.

## Caddy HTTPS not working
- Confirm DNS points to the VPS.
- Ports 80/443 must be open.
- Ensure `DOMAIN` and `ACME_EMAIL` are set in `.env`.
