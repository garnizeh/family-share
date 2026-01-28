# Task 125: Deployment Guide with Docker Compose + Caddy

**Milestone:** Deploy & Docs  
**Points:** 1 (5 hours)  
**Dependencies:** 110  
**Branch:** `feat/deployment`  
**Labels:** `ops`, `documentation`

## Description
Create deployment documentation for running FamilyShare on a VPS using Docker Compose and Caddy.

## Acceptance Criteria
- [ ] Docker Compose file created
- [ ] Deployment guide with step-by-step instructions
- [ ] Reverse proxy config example (Caddy)
- [ ] Environment variable configuration documented
- [ ] Backup and restore instructions
- [ ] Log rotation configured (Docker)

## Files to Add/Modify
- `.docs/deployment/docker-compose.md` — compose setup
- `.docs/deployment/reverse-proxy.md` — Caddy example
- `.docs/deployment/backup-restore.md` — backup strategies
- `deploy/docker-compose.yml` — compose stack
- `deploy/Caddyfile` — reverse proxy config

## Docker Compose (Summary)
```yaml
services:
    app:
        image: ghcr.io/your-org/familyshare:latest
        env_file: .env
        volumes:
            - ./data:/app/data
        restart: unless-stopped
    caddy:
        image: caddy:2
        ports:
            - "80:80"
            - "443:443"
        volumes:
            - ./deploy/Caddyfile:/etc/caddy/Caddyfile:ro
            - caddy_data:/data
            - caddy_config:/config
```

## Caddy Reverse Proxy Example
```
familyshare.example.com {
    reverse_proxy localhost:8080
    encode gzip
    
    header {
        Strict-Transport-Security "max-age=31536000;"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "DENY"
    }
}
```


## Deployment Steps (Compose)
1. Create directories: `/opt/familyshare/{data,deploy}`
2. Copy `deploy/docker-compose.yml` and `deploy/Caddyfile`
3. Configure `.env` file
4. Start: `docker compose up -d`
5. Test access via HTTPS

## Backup Strategy
- **Database**: `sqlite3 familyshare.db ".backup backup.db"` (daily cron)
- **Photos**: `tar -czf photos-backup.tar.gz data/photos/` (daily cron)
- **Environment**: backup `.env` securely (not in repo)

## Tests Required
- [ ] Manual test: containers start and stop correctly
- [ ] Manual test: service restarts on failure
- [ ] Manual test: reverse proxy forwards requests correctly

## PR Checklist
- [ ] Deployment guide is clear and complete
- [ ] Reverse proxy example tested (Caddy)
- [ ] Backup commands tested and verified
- [ ] Security hardening options documented

## Git Workflow
```bash
git checkout -b feat/deployment
# Create deployment docs and compose stack
# Test on VPS or VM
git add .docs/deployment/ deploy/
git commit -m "docs: add deployment guide and compose stack"
git push origin feat/deployment
# Open PR: "Add deployment documentation with Caddy + Compose"
```

## Notes
- Test deployment on a fresh VPS (Digital Ocean, Hetzner, etc.)
- Document port 8080 (or make configurable)
- Include firewall rules (ufw/iptables)
- Recommend Caddy for simplicity (auto-HTTPS)
