# Quick Deploy Reference

One-page reference for deploying FamilyShare to a VPS.

## Prerequisites
- Ubuntu 20.04+ VPS (1-2GB RAM)
- Domain pointing to VPS IP
- SSH access

## 1-Minute Deploy

```bash
# On VPS
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# Logout and login again

git clone https://github.com/YOUR_USERNAME/family-share.git
cd family-share

# Configure
# Use the repository's example .env (most repos provide `.env.example`)
cp .env.example .env
nano .env  # Edit DOMAIN, ACME_EMAIL, and generate secrets

# Generate admin password
go run scripts/hash_password.go YourSecurePassword123
# Copy hash to .env ADMIN_PASSWORD_HASH

# Generate secrets
openssl rand -hex 32  # For CSRF_SECRET
openssl rand -hex 32  # For VIEWER_HASH_SECRET

# Deploy
./scripts/deploy.sh
```

## Configuration Checklist

In `.env`:
- ✅ `ADMIN_PASSWORD_HASH` - Generate with `hash_password.go`
- ✅ `CSRF_SECRET` - Generate with `openssl rand -hex 32`
- ✅ `VIEWER_HASH_SECRET` - Generate with `openssl rand -hex 32`
- ✅ `DOMAIN` - Your domain name
- ✅ `ACME_EMAIL` - Your email for Let's Encrypt
- ✅ `FORCE_HTTPS=true` - Enable for production
- ✅ `APP_ENV=production` - Production mode

## Essential Commands

```bash
# Navigate to project
cd ~/family-share/deploy

# View logs
docker compose logs -f

# Restart
docker compose restart

# Stop
docker compose down

# Update and redeploy
cd ..
git pull
./scripts/deploy.sh

# Backup database
cp data/familyshare.db backups/backup-$(date +%Y%m%d).db

# Check disk usage
du -sh data/
docker system df
```

## Firewall Setup

```bash
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable
```

## Troubleshooting

**App won't start:**
```bash
docker compose logs app
# Check .env file for errors
```

**SSL not working:**
```bash
docker compose logs caddy
# Verify domain DNS: dig your-domain.com
```

**Out of space:**
```bash
docker system prune -a -f
du -sh data/photos/
```

## Health Check

```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy","timestamp":"..."}
```

## Access Points

- Admin: `https://your-domain.com/admin/login`
- Health: `https://your-domain.com/health`
- Share links: `https://your-domain.com/s/{token}`

## Monitoring

```bash
# Container stats
docker stats

# Disk usage
df -h
du -sh ~/family-share/data

# Active connections
docker compose exec app netstat -tunlp
```

## Backup Automation

Add to crontab:
```bash
crontab -e

# Daily backup at 2 AM
0 2 * * * cp ~/family-share/data/familyshare.db ~/backups/familyshare-$(date +\%Y\%m\%d).db

# Weekly cleanup (keep last 7 days)
0 3 * * 0 find ~/backups -name "familyshare-*.db" -mtime +7 -delete
```

## Performance Tips

1. Use WebP instead of AVIF (faster)
2. Set `IMAGE_FORMAT=webp` in .env
3. Limit concurrent uploads (future feature)
4. Regular cleanup: `docker system prune -a`
5. Monitor with: `htop` and `docker stats`

## Security Checklist

- ✅ Strong admin password (20+ chars)
- ✅ Random CSRF and viewer hash secrets
- ✅ FORCE_HTTPS=true
- ✅ Firewall enabled
- ✅ SSH key authentication
- ✅ Regular updates: `apt-get update && apt-get upgrade`
- ✅ Automated backups
- ✅ Rate limiting enabled

## Full Documentation

For detailed instructions, see:
- [VPS Deployment Guide](.docs/deployment/vps-deployment.md)
- [Configuration Guide](.docs/configuration.md)
- [Troubleshooting](.docs/troubleshooting.md)
