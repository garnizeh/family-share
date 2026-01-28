# FamilyShare

Lightweight, self-hosted photo sharing for families on low-resource VPS.

This repository contains the Go backend and SSR frontend. See `.docs/` for design notes and task backlog.

## Features
- Zero-waste storage: uploads are resized and converted to WebP
- Magic link sharing with view/time limits
- Simple admin UI (no SPA)
- Mobile-first UX
- HTMX + Alpine.js for light interactivity

## Tech Stack
- Go (`net/http`, `chi`)
- SQLite (pure Go, no CGO)
- HTML templates + Tailwind CSS

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
- Build from source (see Quick Start)
- VPS deployment (Docker Compose + Caddy): see `.docs/deployment/docker-compose.md`

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

**Q: Do viewers need an account?**
A: No. Access is via share links only.

## Database migrations
Migrations are embedded in the binary under `sql/schema/*` and applied at startup by `internal/db`.

Manual apply (debug only):

```bash
sqlite3 ./data/familyshare.db < sql/schema/0001_init_schema.sql
```
