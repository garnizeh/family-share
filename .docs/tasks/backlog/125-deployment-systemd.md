# Task 125: Deployment Guide and Systemd Service

**Milestone:** Deploy & Docs  
**Points:** 1 (5 hours)  
**Dependencies:** 110  
**Branch:** `feat/deployment`  
**Labels:** `ops`, `documentation`

## Description
Create deployment documentation and systemd service file for running FamilyShare on a VPS. Include reverse proxy configuration examples.

## Acceptance Criteria
- [ ] Systemd service file created (`.service`)
- [ ] Deployment guide with step-by-step instructions
- [ ] Reverse proxy config examples (Caddy, Nginx)
- [ ] Environment variable configuration documented
- [ ] Backup and restore instructions
- [ ] Log rotation configured

## Files to Add/Modify
- `.docs/deployment/systemd-service.md` — service file and setup
- `.docs/deployment/reverse-proxy.md` — Caddy/Nginx examples
- `.docs/deployment/backup-restore.md` — backup strategies
- `deploy/familyshare.service` — systemd unit file

## Systemd Service File
```ini
[Unit]
Description=FamilyShare Photo Sharing Service
After=network.target

[Service]
Type=simple
User=familyshare
Group=familyshare
WorkingDirectory=/opt/familyshare
EnvironmentFile=/opt/familyshare/.env
ExecStart=/opt/familyshare/bin/familyshare
Restart=on-failure
RestartSec=5s

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/opt/familyshare/data

[Install]
WantedBy=multi-user.target
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

## Nginx Reverse Proxy Example
```nginx
server {
    listen 443 ssl http2;
    server_name familyshare.example.com;
    
    ssl_certificate /etc/letsencrypt/live/familyshare.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/familyshare.example.com/privkey.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Deployment Steps
1. Build binary: `go build -o familyshare ./cmd/familyshare`
2. Create user: `useradd -r -s /bin/false familyshare`
3. Create directories: `/opt/familyshare/{bin,data}`
4. Copy binary and set permissions
5. Configure `.env` file
6. Install systemd service
7. Enable and start: `systemctl enable --now familyshare`
8. Configure reverse proxy (Caddy or Nginx)
9. Test access via HTTPS

## Backup Strategy
- **Database**: `sqlite3 familyshare.db ".backup backup.db"` (daily cron)
- **Photos**: `tar -czf photos-backup.tar.gz data/photos/` (daily cron)
- **Environment**: backup `.env` securely (not in repo)

## Tests Required
- [ ] Manual test: systemd service starts and stops correctly
- [ ] Manual test: service restarts on failure
- [ ] Manual test: reverse proxy forwards requests correctly

## PR Checklist
- [ ] Systemd service file tested on Linux
- [ ] Deployment guide is clear and complete
- [ ] Reverse proxy examples tested (at least one)
- [ ] Backup commands tested and verified
- [ ] Security hardening options documented

## Git Workflow
```bash
git checkout -b feat/deployment
# Create deployment docs and service file
# Test on VPS or VM
git add .docs/deployment/ deploy/
git commit -m "docs: add deployment guide and systemd service"
git push origin feat/deployment
# Open PR: "Add deployment documentation and systemd service"
```

## Notes
- Test deployment on a fresh VPS (Digital Ocean, Hetzner, etc.)
- Document port 8080 (or make configurable)
- Include firewall rules (ufw/iptables)
- Recommend Caddy for simplicity (auto-HTTPS)
