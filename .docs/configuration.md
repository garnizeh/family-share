# Configuration Reference

FamilyShare reads configuration from environment variables (optionally via `.env`).

| Variable | Default | Description |
| --- | --- | --- |
| `SERVER_ADDR` | `:8080` | Address the HTTP server listens on. |
| `DATABASE_PATH` | `./data/familyshare.db` | SQLite database file path. |
| `DATA_DIR` | `./data` | Base directory for stored photos and assets. |
| `STORAGE_PATH` | `./data` | Storage path used by the image pipeline (set this to match `DATA_DIR`). |
| `TEMP_UPLOAD_DIR` | system temp | Directory for temporary upload files. |
| `ADMIN_PASSWORD_HASH` | empty | bcrypt hash for admin login. |
| `RATE_LIMIT_SHARE` | `60` | Requests/min for public share links. |
| `RATE_LIMIT_ADMIN` | `10` | Requests/min for admin endpoints. |
| `TRUSTED_PROXY_CIDRS` | empty | Comma-separated CIDR ranges for trusted proxies (honor forwarded headers only when the request originates from these ranges). |
| `JANITOR_INTERVAL` | `6h` | Cleanup interval for expired links/files. |
| `DOMAIN` | none | Caddy site domain (Compose deployment). |
| `ACME_EMAIL` | none | Email for ACME/TLS registration in Caddy. |

## Password Hash
Generate a bcrypt hash for the admin password:

```
make hash-password PASSWORD=YourSecurePassword123
```

or

```
go run scripts/hash_password.go YourSecurePassword123
```

Set the result in `ADMIN_PASSWORD_HASH`.

## Notes about `DOMAIN` / `ACME_EMAIL`

- These values are used by the project's deployment workflow (see `.docs/deployment/vps-deployment.md`). The deploy script validates that `DOMAIN` and `ACME_EMAIL` are present in your project `.env` and will bring up Caddy so TLS can be provisioned automatically.
- Ensure your DNS for `DOMAIN` points to the VPS before running the deploy script so ACME can complete.
