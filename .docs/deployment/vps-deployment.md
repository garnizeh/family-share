# VPS Deployment Guide

This guide walks you through deploying FamilyShare to a VPS using Docker Compose and Caddy for automatic HTTPS.

## Prerequisites

- VPS with Ubuntu 20.04+ (1GB RAM minimum, 2GB recommended)
- Domain name pointing to your VPS IP
- SSH access to your VPS

## Step 1: Prepare Your VPS

```bash
# SSH into your VPS
ssh user@your-vps-ip

# Update system
sudo apt-get update && sudo apt-get upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo apt-get install docker-compose-plugin -y

# Add your user to docker group (logout/login required after this)
sudo usermod -aG docker $USER

# Verify installations
docker --version
docker compose version
```

## Step 2: Clone Repository on VPS

```bash
# Create app directory
mkdir -p ~/apps
cd ~/apps

# Clone repository
git clone https://github.com/YOUR_USERNAME/family-share.git
cd family-share
```

## Step 3: Configure Environment

```bash
# Create .env file from example
cp .env.example .env

# Generate admin password hash locally or on VPS
go run scripts/hash_password.go YourSecurePassword123

# Edit .env file
nano .env
```

Update these critical values in `.env`:

```bash
# Server configuration
SERVER_ADDR=:8080
DATA_DIR=/app/data
DATABASE_PATH=/app/data/familyshare.db
TEMP_UPLOAD_DIR=/app/tmp_uploads

# Admin credentials
ADMIN_PASSWORD_HASH='$2a$12$YOUR_GENERATED_HASH_HERE'

# Image format (webp or avif)
IMAGE_FORMAT=webp

# Rate limiting
RATE_LIMIT_SHARE=60
RATE_LIMIT_ADMIN=10

# Optional: Trusted proxy (if behind Cloudflare)
# TRUSTED_PROXY=cloudflare
```

## Step 4: Configure domain and TLS via .env

You do NOT need to edit `deploy/Caddyfile` for most deployments — the deploy script reads `DOMAIN` and `ACME_EMAIL` from your project `.env` and will validate they exist before bringing up Caddy. This keeps deployment reproducible and avoids manual edits to the checked-in Caddyfile.

Open your project `.env` (created in Step 3) and set these two values:

```bash
# At project root .env
DOMAIN=photos.yourdomain.com
ACME_EMAIL=your-email@example.com
```

Notes:
- If you need custom Caddy configuration (advanced users), you can still edit `deploy/Caddyfile` directly — but this is optional. The default Caddyfile included with the project is suitable for almost all VPS deployments.
- Ensure DNS for `DOMAIN` points to your VPS before running the deploy script so Caddy can provision certificates.

## Step 5: Build and Deploy

Everything is now handled by the deployment script. From the project root run:

```bash
# Run the deploy script (default: production). The script performs git pull,
# validates .env, builds the Docker image, starts services, runs health checks
# and prunes old images.
./scripts/deploy.sh [production|staging]
```

Notes:
- The script expects a configured `.env` file at the project root (see Step 3).
- It runs from `deploy/` internally and uses `docker compose` there, so you do
   not need to run `docker compose` manually.
- The script will prompt for confirmation before making changes and will
   create a database backup if an existing database is found.

Typical final output shows services running, for example:

```
NAME                    IMAGE               STATUS
deploy-app-1           familyshare:latest   Up
deploy-caddy-1         caddy:2             Up
```

## Step 6: Verify Deployment

1. **Check app health:**
   ```bash
   curl http://localhost:8080/health
   ```

2. **Access admin panel:**
   - Navigate to `https://your-domain.com/admin/login`
   - Login with your admin password

3. **Check SSL certificate:**
   - Caddy automatically provisions Let's Encrypt certificates
   - Check browser for valid HTTPS

## Step 7: Setup Firewall (Optional but Recommended)

```bash
# Install UFW if not present
sudo apt-get install ufw -y

# Allow SSH (IMPORTANT: do this first!)
sudo ufw allow 22/tcp

# Allow HTTP and HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# Enable firewall
sudo ufw enable

# Check status
sudo ufw status
```

## Ongoing Maintenance

### View Logs

```bash
cd ~/apps/family-share/deploy

# All logs
docker compose logs -f

# App only
docker compose logs -f app

# Caddy only
docker compose logs -f caddy
```

### Update Application

```bash
cd ~/apps/family-share

# Pull latest changes
git pull origin main

# Rebuild and restart
cd deploy
docker compose down
docker compose up -d --build

# Clean old images
docker image prune -f
```

Prefer using the deployment script which automates pull, build, backup and restart:

```bash
# From project root
./scripts/deploy.sh [production|staging]
```

### Backup Database

```bash
# Create backup directory
mkdir -p ~/backups

# Backup database
cp ~/apps/family-share/data/familyshare.db ~/backups/familyshare-$(date +%Y%m%d-%H%M%S).db

# Or use docker cp
docker compose exec app sqlite3 /app/data/familyshare.db ".backup '/app/data/backup.db'"
```

### Restore Database

```bash
# Stop app
cd ~/apps/family-share/deploy
docker compose stop app

# Restore from backup
cp ~/backups/familyshare-YYYYMMDD-HHMMSS.db ~/apps/family-share/data/familyshare.db

# Start app
docker compose start app
```

### Monitor Disk Usage

```bash
# Check data directory size
du -sh ~/apps/family-share/data

# Check Docker disk usage
docker system df

# Clean up unused Docker resources
docker system prune -a
```

## Troubleshooting

### App won't start

```bash
# Check logs
docker compose logs app

# Common issues:
# 1. Invalid ADMIN_PASSWORD_HASH in .env
# 2. Port 8080 already in use
# 3. Permission issues with data directory
```

### Caddy SSL certificate issues

```bash
# Check Caddy logs
docker compose logs caddy

# Verify DNS is pointing to your VPS
dig your-domain.com

# Ensure ports 80 and 443 are open
sudo ufw status
```

### Database locked errors

```bash
# Check if multiple instances are running
docker compose ps

# Restart app
docker compose restart app
```

### Out of disk space

```bash
# Check disk usage
df -h

# Clean old photos (careful!)
# Photos are stored in: ~/apps/family-share/data/photos/

# Clean Docker
docker system prune -a -f
```

## Alternative: Deploy Without Docker

If you prefer running without Docker:

```bash
# Install build dependencies
sudo apt-get install -y gcc libc6-dev

# Build binary
cd ~/apps/family-share
go build -o familyshare ./cmd/app

# Create systemd service
sudo nano /etc/systemd/system/familyshare.service
```

Add:

```ini
[Unit]
Description=FamilyShare Photo Sharing
After=network.target

[Service]
Type=simple
User=YOUR_USERNAME
WorkingDirectory=/home/YOUR_USERNAME/apps/family-share
EnvironmentFile=/home/YOUR_USERNAME/apps/family-share/.env
ExecStart=/home/YOUR_USERNAME/apps/family-share/familyshare
Restart=on-failure
RestartSec=5s

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable familyshare
sudo systemctl start familyshare
sudo systemctl status familyshare

# Setup Caddy separately
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy

# Copy Caddyfile
sudo cp ~/apps/family-share/deploy/Caddyfile /etc/caddy/Caddyfile
sudo systemctl restart caddy
```

## Security Best Practices

1. **Keep system updated:**
   ```bash
   sudo apt-get update && sudo apt-get upgrade -y
   ```

2. **Setup automatic security updates:**
   ```bash
   sudo apt-get install unattended-upgrades -y
   sudo dpkg-reconfigure -plow unattended-upgrades
   ```

3. **Use strong admin password** (20+ characters)

4. **Setup SSH key authentication** and disable password login

5. **Regular backups** (automate with cron)

6. **Monitor logs** for suspicious activity

7. **Limit rate limiting** is already configured in the app

## Performance Optimization

For low-resource VPS (1GB RAM):

1. **Limit Docker memory:**
   ```yaml
   # In docker-compose.yml
   services:
     app:
       mem_limit: 512m
   ```

2. **Reduce concurrent uploads** by setting smaller upload limits

3. **Use WebP instead of AVIF** (faster encoding)

4. **Setup log rotation** (already configured in docker-compose.yml)

5. **Monitor with htop:**
   ```bash
   sudo apt-get install htop -y
   htop
   ```

## Next Steps

- Setup automated backups (see backup section)
- Configure monitoring (e.g., Uptime Robot)
- Setup email notifications (future feature)
- Consider adding rate limiting at Caddy level for additional protection
