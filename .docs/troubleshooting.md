# Troubleshooting

## Login fails
- Ensure `ADMIN_PASSWORD_HASH` is set to a bcrypt hash (not plain text).
- Regenerate hash using `make hash-password`.

## Photos not uploading
- Verify `STORAGE_PATH`/`DATA_DIR` are writable.
- Check disk space on the VPS.
- If using Docker, ensure volumes are mounted correctly.

## Multipart "NextPart" i/o timeouts (uploads aborting)

- Symptoms: logs show repeated messages like `multipart read error: multipart: NextPart: read tcp ...: i/o timeout`.
- Causes: the client is streaming slowly or the connection is paused long enough that either the HTTP server or the reverse proxy closes the connection due to a read/write timeout.
- Mitigation:
	- Ensure the reverse proxy (Caddy/nginx) read/write timeouts are set high enough for large uploads. When using the project's `./scripts/deploy.sh` and the included `deploy/Caddyfile`, the transport timeouts are already tuned for larger uploads. If you use a custom proxy, increase `read_timeout` / `write_timeout`.
	- Make sure the host HTTP server's read timeout is not too small. If you run FamilyShare directly (no container), check your systemd or process manager socket timeouts.
	- Prefer using the provided deployment script which configures Caddy and runs health checks.

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
