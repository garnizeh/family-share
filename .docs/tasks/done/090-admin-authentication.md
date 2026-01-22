# Task 090: Security — Admin Authentication and Sessions

**Milestone:** Security & Ops  
**Points:** 2 (7 hours)  
**Dependencies:** 085  
**Branch:** `feat/admin-auth`  
**Labels:** `security`, `admin`, `authentication`

## Description
Implement simple password-based admin authentication with secure session management. Store sessions in SQLite.

## Acceptance Criteria
- [ ] `GET /admin/login` — login form
- [ ] `POST /admin/login` — authenticate and create session
- [ ] `POST /admin/logout` — destroy session
- [ ] Password verified against bcrypt hash from environment
- [ ] Session cookie is HttpOnly, Secure (in prod), SameSite=Lax
- [ ] Auth middleware protects all `/admin/*` routes except `/admin/login`

## Files to Add/Modify
- `internal/handler/admin_auth.go` — login/logout handlers
- `internal/middleware/auth.go` — session authentication middleware
- `internal/security/password.go` — password hashing and verification
- `migrations/0002_add_sessions.sql` — sessions table
- `web/templates/admin/login.html` — login form

## Sessions Table
```sql
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

## Password Verification
```go
func VerifyAdminPassword(providedPassword, envHash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(envHash), []byte(providedPassword))
    return err == nil
}
```

## Session Middleware
```go
func RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("session_id")
        if err != nil {
            http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
            return
        }
        
        session := queries.GetSessionByID(cookie.Value)
        if session == nil || time.Now().After(session.ExpiresAt) {
            http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
```

## Tests Required
- [ ] Integration test: login with correct password creates session
- [ ] Integration test: login with wrong password fails
- [ ] Integration test: accessing /admin/* without session redirects to login
- [ ] Integration test: logout destroys session
- [ ] Unit test: password verification with bcrypt

## PR Checklist
- [ ] Password hash loaded from environment (ADMIN_PASSWORD_HASH)
- [ ] Session expiration enforced (e.g., 24h)
- [ ] Cookie settings are secure (HttpOnly, Secure in prod)
- [ ] All admin routes protected by middleware
- [ ] Tests pass: `go test ./internal/handler/... ./internal/middleware/... -v`

## Git Workflow
```bash
git checkout -b feat/admin-auth
# Implement authentication
go test ./internal/... -v
git add internal/ migrations/ web/templates/admin/
git commit -m "feat: implement admin authentication and session management"
git push origin feat/admin-auth
# Open PR: "Add admin authentication with secure sessions"
```

## Notes
- For MVP, single admin user (no user management)
- Generate ADMIN_PASSWORD_HASH with: `bcrypt -c 12 mypassword`
- Session cleanup handled by janitor (task 110)
- Consider CSRF protection for state-changing endpoints (next task)
