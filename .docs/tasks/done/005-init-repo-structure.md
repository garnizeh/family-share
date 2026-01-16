# Task 005: Initialize Repository Structure

**Milestone:** Setup  
**Points:** 1 (4 hours)  
**Dependencies:** None  
**Branch:** `feat/init-repo`  
**Labels:** `setup`, `infrastructure`

## Description
Set up the initial project structure with Go module, environment configuration, and folder layout following Go best practices and the FamilyShare TDD.

## Acceptance Criteria
- [ ] Go module initialized with `familyshare` name
- [ ] `.env.example` created with all required configuration keys
- [ ] `.gitignore` configured for Go projects (binaries, `.env`, `data/`, etc.)
- [ ] Folder structure created: `cmd/`, `internal/`, `web/`, `migrations/`
- [ ] Minimal `README.md` with project description and quick-start placeholder

## Files to Add/Modify
- `go.mod` — already exists, verify correct module name
- `.env.example` — all environment variables with comments
- `.gitignore` — Go-standard ignore patterns + project-specific
- `cmd/app/main.go` — empty main entry point
- `internal/` — create directory
- `internal/db/` — create directory
- `internal/storage/` — create directory
- `internal/handler/` — create directory
- `web/templates/` — create directory
- `web/static/` — create directory
- `migrations/` — create directory for SQL files
- `README.md` — project overview, tech stack, setup instructions (placeholder)

## Tests Required
- None (structural setup)

## PR Checklist
- [x] All directories created and committed (use `.gitkeep` for empty dirs)
- [x] `.env.example` has comments for each variable
- [x] `.gitignore` prevents committing secrets and build artifacts
- [x] `go mod tidy` runs without errors
- [x] README contains project name, description, and basic structure

## Git Workflow
```bash
git checkout -b feat/init-repo
# Create files and directories
git add .
git commit -m "feat: initialize project structure and config"
git push origin feat/init-repo
# Open PR with title: "Initialize repository structure and configuration"
```

## Environment Variables (.env.example)
```
# Admin credentials
ADMIN_PASSWORD_HASH=

# Security
SESSION_SECRET=
TOKEN_PEPPER=

# Server
BASE_URL=http://localhost:8080
PORT=8080

# Storage
STORAGE_PATH=./data

# Rate limiting
RATE_LIMIT_PER_MIN=60

# Cookies (set to true in production with HTTPS)
COOKIE_SECURE=false
```

## Notes
- Use `internal/` for unexported packages (DB, storage, handlers)
- Keep `cmd/app/` minimal; business logic goes in `internal/`
- Embed `web/` assets using `embed.FS` in future tasks
