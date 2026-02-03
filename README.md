# FamilyShare

Lightweight, self-hosted photo sharing for families on low-resource VPS.

This repository contains the Go backend and SSR frontend. See `.docs/` for design notes and task backlog.

## Features
- Zero-waste storage: uploads are resized and converted to WebP (default) or AVIF (optional)
- Magic link sharing with view/time limits
- Simple admin UI (no SPA)
- Mobile-first UX
- HTMX + Alpine.js for light interactivity

## Tech Stack
- Go (`net/http`, `chi`)
- SQLite (pure Go, no CGO)
- HTML templates + Tailwind CSS

## Image Encoders
FamilyShare uses a lightweight image pipeline designed for low-resource servers:
- **WebP** (default): encoded with `github.com/chai2010/webp` at 80% quality.
- **AVIF** (optional): encoded with `github.com/gen2brain/avif` (CGO-free, WASM fallback) using quality 60 and speed 6.

To enable AVIF output, set `IMAGE_FORMAT=avif` in your environment (see `.env.example`). If unset, the pipeline defaults to WebP.

## Quick Start (Local)
1. Copy `.env.example` to `.env` and edit values.
2. Generate an admin password hash:

```bash
make hash-password PASSWORD=YourSecurePassword123
```

Or:

```bash
go run scripts/hash_password.go YourSecurePassword123
```

3. Set the hash:

```bash
export ADMIN_PASSWORD_HASH='$2a$12$...'
```

4. Build and run:

```bash
go build -o familyshare ./cmd/app
./familyshare
```

Admin UI: `http://localhost:8080/admin/login`

## Installation

### Local Development
See Quick Start section above.

### VPS Deployment

**Quick Deploy:**
```bash
# On your VPS
git clone https://github.com/YOUR_USERNAME/family-share.git
cd family-share
cp .env.example .env
nano .env  # Configure your settings
./scripts/deploy.sh
```

**Detailed Guide:**
See [VPS Deployment Guide](.docs/deployment/vps-deployment.md) for complete instructions including:
- VPS preparation
- Domain configuration
- SSL setup with Caddy
- Backup strategies
- Troubleshooting

## Configuration
See `.docs/configuration.md` for all environment variables and defaults.

## Usage
See `.docs/usage-guide.md` for admin workflows (albums, uploads, share links).

## Deployment
See `.docs/deployment/docker-compose.md` and `.docs/deployment/reverse-proxy.md`.

## Troubleshooting
See `.docs/troubleshooting.md`.

## FAQ
**Q: Are original photos stored?**
A: No. Photos are resized and converted to WebP to save space.

**Q: Can I use AVIF instead of WebP?**
A: Yes. Set `IMAGE_FORMAT=avif` to encode uploads as AVIF. WebP remains the default.

**Q: Do viewers need an account?**
A: No. Access is via share links only.

## Database migrations
Migrations are embedded in the binary under `sql/schema/*` and applied at startup by `internal/db`.

Manual apply (debug only):

```bash
sqlite3 ./data/familyshare.db < sql/schema/0001_init_schema.sql
```
