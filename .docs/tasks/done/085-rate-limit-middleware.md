# Task 085: Security — Rate Limiting Middleware

**Milestone:** Security & Ops  
**Points:** 1 (5 hours)  
**Dependencies:** 080  
**Branch:** `feat/rate-limit`  
**Labels:** `security`, `middleware`

## Description
Implement IP-based rate limiting middleware to prevent brute-force attacks on share link tokens and admin endpoints.

## Acceptance Criteria
- [ ] Rate limiter tracks requests per IP per endpoint
- [ ] Configurable limits (e.g., 60 req/min for shares, 10 req/min for admin login)
- [ ] Returns `429 Too Many Requests` when limit exceeded
- [ ] Uses in-memory sliding window or token bucket algorithm
- [ ] Optional: temporary lockout after repeated violations

## Files to Add/Modify
- `internal/middleware/ratelimit.go` — rate limiting middleware
- `internal/middleware/ratelimit_test.go` — unit tests

## Middleware Function
```go
// RateLimit creates middleware with specified limit
func RateLimit(requestsPerMinute int) func(http.Handler) http.Handler

// Apply to routes
r.Group(func(r chi.Router) {
    r.Use(middleware.RateLimit(60))
    r.Get("/s/{token}", handler.ViewShareLink)
})
```

## Implementation Options
- **Simple:** In-memory map with per-IP counters and expiry
- **Library:** Use `golang.org/x/time/rate` for token bucket

## Tests Required
- [ ] Unit test: allows requests under limit
- [ ] Unit test: blocks requests over limit
- [ ] Unit test: limit resets after time window
- [ ] Unit test: different IPs tracked separately
- [ ] Integration test: 429 returned when limit exceeded

## PR Checklist
- [ ] Rate limits configurable via environment variables
- [ ] Memory usage bounded (old entries expire)
- [ ] Works correctly with reverse proxy (X-Forwarded-For header)
- [ ] Tests pass: `go test ./internal/middleware/... -v`
- [ ] Middleware applied to sensitive routes only

## Git Workflow
```bash
git checkout -b feat/rate-limit
# Implement rate limiting middleware
go test ./internal/middleware/... -v -cover
git add internal/middleware/
git commit -m "feat: add rate limiting middleware for security"
git push origin feat/rate-limit
# Open PR: "Implement rate limiting to prevent brute-force"
```

## Notes
- For MVP, simple in-memory rate limiter is sufficient
- Consider using Redis for distributed rate limiting (post-MVP)
- Apply stricter limits to admin login endpoint
- Log rate limit violations for monitoring
- Use `X-RateLimit-*` headers to inform clients
