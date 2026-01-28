# Task 130: README and User Documentation

**Milestone:** Deploy & Docs  
**Points:** 1 (4 hours)  
**Dependencies:** 125  
**Branch:** `feat/readme`  
**Labels:** `documentation`

## Description
Write comprehensive README and user-facing documentation covering installation, configuration, usage, and troubleshooting.

## Acceptance Criteria
- [ ] README includes: project description, features, tech stack, quick-start
- [ ] Installation guide (build from source, download binary)
- [ ] Configuration reference (all environment variables)
- [ ] Usage guide: creating albums, uploading, sharing
- [ ] Troubleshooting section (common issues)
- [ ] FAQ section
- [ ] Contributing guidelines (optional for MVP)

## Files to Add/Modify
- `README.md` — main project README
- `.docs/configuration.md` — environment variable reference
- `.docs/usage-guide.md` — user guide for admins
- `.docs/troubleshooting.md` — common issues and solutions

## README Structure
```markdown
# FamilyShare

Lightweight, self-hosted photo sharing for families on low-resource VPS.

## Features
- Zero-waste storage (WebP conversion, no originals saved)
- Magic link sharing with view/time limits
- Simple admin interface (no complex setup)
- Mobile-first, accessible UX
- HTMX + Alpine.js (no heavy JS frameworks)

## Tech Stack
- Go (net/http, chi router)
- SQLite (pure-Go, no CGO)
- HTMX, Alpine.js, TailwindCSS

## Quick Start
1. Download binary or build from source
2. Configure `.env` (see Configuration Guide)
3. Run: `./familyshare`
4. Access admin at `http://localhost:8080/admin/login`

## Installation
[Link to detailed installation docs]

## Configuration
[Link to configuration reference]

## Deployment
[Link to deployment guide]

## License
MIT (or chosen license)
```

## Configuration Reference
Document all environment variables:
- `ADMIN_PASSWORD_HASH` — bcrypt hash of admin password
- `SESSION_SECRET` — secret for session cookies
- `BASE_URL` — public URL of the app
- `STORAGE_PATH` — path to photo storage
- `PORT` — HTTP port (default 8080)
- `RATE_LIMIT_PER_MIN` — rate limit for public endpoints

## Usage Guide
- How to log in as admin
- Creating albums
- Uploading photos (drag-and-drop)
- Creating share links
- Managing share links (revoke, view stats)
- Setting view limits and expiration

## Troubleshooting
- "Login fails" → Check password hash generation
- "Photos not uploading" → Check storage permissions
- "Share link expired" → Explain expiration logic
- "High disk usage" → Run janitor manually, check VACUUM

## Tests Required
- [ ] Manual test: follow README quick-start from scratch
- [ ] Verify all links in docs are valid
- [ ] Spell-check and grammar review

## PR Checklist
- [ ] README is clear and concise
- [ ] All major features documented
- [ ] Installation steps tested on fresh system
- [ ] Configuration examples provided
- [ ] No sensitive data in examples

## Git Workflow
```bash
git checkout -b feat/readme
# Write documentation
# Test quick-start on clean VM
git add README.md .docs/
git commit -m "docs: add comprehensive README and user documentation"
git push origin feat/readme
# Open PR: "Add README and user documentation"
```

## Notes
- Keep README concise; link to detailed docs
- Use screenshots sparingly (keep repo size small)
- Document default values for all config options
- Include link to live demo (if available)
