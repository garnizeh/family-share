#!/bin/bash
# VPS Deployment Script for FamilyShare
# Usage: ./scripts/deploy.sh [production|staging]

set -e

ENV=${1:-production}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo "🚀 FamilyShare Deployment Script"
echo "Environment: $ENV"
echo "Project root: $PROJECT_ROOT"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Pull latest changes (if in git repo)
# Done EARLY to ensure we validate the latest configuration requirements
if [ -d "$PROJECT_ROOT/.git" ]; then
    echo "📥 Pulling latest changes from git..."
    current_dir=$(pwd)
    cd "$PROJECT_ROOT"
    git pull origin main || print_warning "Could not pull from git (continuing anyway)"
    cd "$current_dir"
    print_success "Git pull complete"
fi

# Check if .env exists
if [ ! -f "$PROJECT_ROOT/.env" ]; then
    print_error ".env file not found!"
    echo "Please copy .env.example to .env and configure it first:"
    echo "  cp .env.example .env"
    echo "  nano .env"
    exit 1
fi

# Load .env variables
set -a
source "$PROJECT_ROOT/.env"
set +a
print_success "Loaded configuration from .env"

# Check if running in deploy directory
cd "$PROJECT_ROOT/deploy"

# Pre-flight checks
echo ""
echo "📋 Running pre-flight checks..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker not found. Please install Docker first."
    exit 1
fi
print_success "Docker is installed"

# Check Docker Compose
if ! docker compose version &> /dev/null; then
    print_error "Docker Compose not found. Please install Docker Compose plugin."
    exit 1
fi
print_success "Docker Compose is installed"

# Check if Caddyfile exists
if [ ! -f "Caddyfile" ]; then
    print_error "Caddyfile not found in deploy directory"
    exit 1
fi

# Check for DOMAIN and ACME_EMAIL in env
# We check irrespective of Caddyfile content to ensure they are available
if grep -q '{$DOMAIN}' Caddyfile || [ -n "$DOMAIN" ]; then
    if [ -z "$DOMAIN" ]; then
        print_error "DOMAIN environment variable not set in .env"
        echo "Please add DOMAIN=yourdomain.com to your .env file."
        exit 1
    fi
    
    if grep -q '{$ACME_EMAIL}' Caddyfile || [ -n "$ACME_EMAIL" ]; then
        if [ -z "$ACME_EMAIL" ]; then
            print_error "ACME_EMAIL environment variable not set in .env"
            echo "Please add ACME_EMAIL=your-email@example.com to your .env file."
            exit 1
        fi
    fi
    
    print_success "Domain: $DOMAIN"
    print_success "Email: $ACME_EMAIL"
fi

# Ask for confirmation
echo ""
echo "🔍 Deployment Summary:"
echo "  - Environment: $ENV"
echo "  - Docker Compose file: docker-compose.yml"
echo "  - Domain: ${DOMAIN:-'configured in Caddyfile'}"
echo ""
read -p "Continue with deployment? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    print_warning "Deployment cancelled"
    exit 0
fi

# Backup database if it exists
if [ -f "$PROJECT_ROOT/data/familyshare.db" ]; then
    echo ""
    echo "💾 Backing up database..."
    BACKUP_DIR="$PROJECT_ROOT/backups"
    mkdir -p "$BACKUP_DIR"
    BACKUP_FILE="$BACKUP_DIR/familyshare-$(date +%Y%m%d-%H%M%S).db"
    cp "$PROJECT_ROOT/data/familyshare.db" "$BACKUP_FILE"
    print_success "Database backed up to: $BACKUP_FILE"
fi

# Stop existing containers
echo ""
echo "🛑 Stopping existing containers..."
docker compose down || true
print_success "Containers stopped"

# Build new image
echo ""
echo "🔨 Building new Docker image..."
docker compose build --no-cache
print_success "Image built"

# Start services
echo ""
echo "🚀 Starting services..."
docker compose up -d
print_success "Services started"

# Wait for app to be ready
echo ""
echo "⏳ Waiting for app to be ready..."
sleep 5

# Health check
echo ""
echo "🏥 Running health check..."
MAX_RETRIES=10
RETRY_COUNT=0

# Use localhost health check since we are on the server
while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if docker compose exec -T app wget -q -O - http://localhost:8080/health > /dev/null 2>&1; then
        print_success "Health check passed!"
        break
    fi
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "Retry $RETRY_COUNT/$MAX_RETRIES..."
    sleep 2
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    print_error "Health check failed after $MAX_RETRIES retries"
    echo ""
    echo "📋 Container logs:"
    docker compose logs --tail=50 app
    exit 1
fi

# Show running containers
echo ""
echo "📊 Container status:"
docker compose ps

# Show logs
echo ""
echo "📋 Recent logs:"
docker compose logs --tail=20

# Cleanup old images
echo ""
echo "🧹 Cleaning up old images..."
docker image prune -f > /dev/null 2>&1
print_success "Cleanup complete"

# Final status
echo ""
echo "============================================="
print_success "Deployment completed successfully!"
echo "============================================="
echo ""
echo "📌 Next steps:"
echo "  1. Visit https://${DOMAIN:-your-domain.com}/admin/login"
echo "  2. Login with your admin password"
echo "  3. Monitor logs: cd $PROJECT_ROOT/deploy && docker compose logs -f"
echo ""
echo "Useful commands:"
echo "  - View logs:     docker compose logs -f"
echo "  - Restart:       docker compose restart"
echo "  - Stop:          docker compose down"
echo "  - Update:        ./scripts/deploy.sh"
echo ""
