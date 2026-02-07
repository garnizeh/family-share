# Reverse Proxy (Caddy)

Use the provided `deploy/Caddyfile`. For standard VPS deployments you only need to set these variables in your project `.env` (the deploy script validates them):

```
DOMAIN=familyshare.example.com
ACME_EMAIL=admin@example.com
```

Caddy will automatically obtain and renew TLS certificates.

Important: timeouts and uploads

- Long or slow multipart uploads can be terminated by the reverse proxy if its read/write timeouts are too low. Ensure your reverse proxy timeouts are equal to or greater than the expected upload streaming time. When using the included `deploy/Caddyfile` and the `./scripts/deploy.sh` workflow, the default Caddy transport timeouts are set to accommodate typical large uploads. If you maintain a custom Caddy setup, increase `read_timeout` / `write_timeout` accordingly.

## Health Check
After services are started (via `./scripts/deploy.sh`), confirm:
- `https://familyshare.example.com` loads
- Requests are forwarded to the app container
