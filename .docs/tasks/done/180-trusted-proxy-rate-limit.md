# Task 180: Security — Trusted Proxy IP Handling

**Milestone:** Security & Ops  
**Points:** 2 (5 hours)  
**Dependencies:** 085  
**Branch:** `feat/trusted-proxy-ips`  
**Labels:** `security`, `middleware`

## Description
Honor `X-Forwarded-For` only when the request originates from a trusted proxy range, to avoid IP spoofing.

## Acceptance Criteria
- [ ] Trusted proxy CIDRs configurable via env
- [ ] Forwarded headers ignored when request is not from trusted proxy
- [ ] Rate limiter and viewer hash use the validated client IP

## Files to Add/Modify
- `internal/middleware/ratelimit.go`
- `internal/security/viewer_hash.go`
- `internal/config/config.go`
- `.env.example`

## Implementation Notes
- Add helper to parse CIDR list and validate `r.RemoteAddr`.
- Only then use `X-Forwarded-For` or `X-Real-IP`.

## Tests Required
- [ ] Unit: trusted proxy allows forwarded IP
- [ ] Unit: untrusted proxy ignores forwarded IP

## PR Checklist
- [ ] Defaults safe for local dev
- [ ] Docs updated for proxy settings

## Git Workflow
```bash
git checkout -b feat/trusted-proxy-ips
# Implement trusted proxy handling

go test ./internal/middleware/... -v

go test ./internal/security/... -v

git add internal/ .env.example

git commit -m "feat: honor forwarded IPs only from trusted proxies"

git push origin feat/trusted-proxy-ips
```
