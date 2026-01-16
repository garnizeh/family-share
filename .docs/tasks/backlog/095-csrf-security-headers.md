# Task 095: Security — CSRF Protection and Headers

**Milestone:** Security & Ops  
**Points:** 1 (4 hours)  
**Dependencies:** 090  
**Branch:** `feat/security-headers`  
**Labels:** `security`, `middleware`

## Description
Add CSRF token protection for admin state-changing requests and configure secure HTTP headers (CSP, X-Frame-Options, etc.).

## Acceptance Criteria
- [ ] CSRF tokens generated and validated for POST/DELETE/PUT requests
- [ ] HTMX requests include CSRF token in header or hidden input
- [ ] Security headers applied globally via middleware
- [ ] CSP configured to allow HTMX and Alpine.js CDNs
- [ ] X-Frame-Options, X-Content-Type-Options set

## Files to Add/Modify
- `internal/middleware/csrf.go` — CSRF middleware
- `internal/middleware/headers.go` — security headers middleware
- `web/templates/layouts/base.html` — include CSRF meta tag

## CSRF Implementation
```go
func CSRFProtection(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
            next.ServeHTTP(w, r)
            return
        }
        
        token := r.Header.Get("X-CSRF-Token")
        if token == "" {
            token = r.FormValue("csrf_token")
        }
        
        if !verifyCSRFToken(token, r) {
            http.Error(w, "Invalid CSRF token", http.StatusForbidden)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## Security Headers
```go
func SecurityHeaders(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Security-Policy", 
            "default-src 'self'; script-src 'self' https://unpkg.com; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net")
        w.Header().Set("X-Frame-Options", "DENY")
        w.Header().Set("X-Content-Type-Options", "nosniff")
        w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
        
        next.ServeHTTP(w, r)
    })
}
```

## HTMX CSRF Integration
```html
<meta name="csrf-token" content="{{ .CSRFToken }}">
<script>
document.body.addEventListener('htmx:configRequest', (event) => {
    event.detail.headers['X-CSRF-Token'] = document.querySelector('meta[name="csrf-token"]').content;
});
</script>
```

## Tests Required
- [ ] Unit test: CSRF token validation passes for valid token
- [ ] Unit test: CSRF token validation fails for invalid token
- [ ] Integration test: POST request without CSRF returns 403
- [ ] Integration test: HTMX request with CSRF header succeeds
- [ ] Unit test: security headers applied to all responses

## PR Checklist
- [ ] CSRF tokens generated per session
- [ ] HTMX auto-includes CSRF token in all requests
- [ ] CSP allows required CDNs (HTMX, Alpine, Tailwind)
- [ ] Tests pass: `go test ./internal/middleware/... -v`
- [ ] Manual test: forms submit successfully with CSRF

## Git Workflow
```bash
git checkout -b feat/security-headers
# Implement CSRF and headers
go test ./internal/middleware/... -v
git add internal/middleware/ web/templates/
git commit -m "feat: add CSRF protection and security headers"
git push origin feat/security-headers
# Open PR: "Implement CSRF protection and security headers"
```

## Notes
- CSRF tokens can be stored in session or signed cookies
- For MVP, simple HMAC-based token verification is sufficient
- CSP should be strict but allow necessary CDNs
- Test in browser to ensure HTMX works with CSRF
