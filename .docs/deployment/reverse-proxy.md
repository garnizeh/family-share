# Reverse Proxy (Caddy)

Use the provided `deploy/Caddyfile` and configure these variables in `.env`:

```
DOMAIN=familyshare.example.com
ACME_EMAIL=admin@example.com
```

Caddy will automatically obtain and renew TLS certificates.

## Health Check
After `docker compose up -d`, confirm:
- `https://familyshare.example.com` loads
- Requests are forwarded to the app container
