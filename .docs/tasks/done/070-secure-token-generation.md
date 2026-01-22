# Task 070: Secure Token Generation for Share Links

**Milestone:** Sharing & View Logic  
**Points:** 1 (4 hours)  
**Dependencies:** 020  
**Branch:** `feat/token-gen`  
**Labels:** `security`, `sharing`

## Description
Implement cryptographically secure token generation for share links using `crypto/rand`. Ensure tokens are URL-safe and sufficiently random.

## Acceptance Criteria
- [ ] Tokens generated with `crypto/rand`
- [ ] Token length is 32 bytes (256 bits) minimum
- [ ] Tokens encoded as URL-safe base64
- [ ] Token uniqueness enforced by database constraint
- [ ] Collision retry logic implemented (retry on duplicate)

## Files to Add/Modify
- `internal/security/tokens.go` — token generation
- `internal/security/tokens_test.go` — unit tests

## Key Functions
```go
// GenerateSecureToken creates a cryptographically random URL-safe token
func GenerateSecureToken() (string, error)

// TokenLength is the byte length of raw tokens (before encoding)
const TokenLength = 32
```

## Implementation
```go
func GenerateSecureToken() (string, error) {
    b := make([]byte, TokenLength)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return base64.URLEncoding.EncodeToString(b), nil
}
```

## Tests Required
- [ ] Unit test: token length is correct
- [ ] Unit test: tokens are URL-safe (no +, /, =)
- [ ] Unit test: 1000 tokens are unique (no collisions)
- [ ] Unit test: error handling when rand.Read fails
- [ ] Unit test: tokens have high entropy (basic randomness check)

## PR Checklist
- [ ] `crypto/rand` used (NOT `math/rand`)
- [ ] Tokens are base64 URL-encoded
- [ ] No hardcoded seeds or predictable patterns
- [ ] Tests pass: `go test ./internal/security/... -v`
- [ ] Code reviewed for security best practices

## Git Workflow
```bash
git checkout -b feat/token-gen
# Implement secure token generation
go test ./internal/security/... -v -cover
git add internal/security/
git commit -m "feat: implement secure token generation for share links"
git push origin feat/token-gen
# Open PR: "Add cryptographically secure token generation"
```

## Notes
- 32 bytes = 256 bits of entropy (sufficient for share links)
- URL-safe base64 avoids issues with `/` and `+` in URLs
- Database UNIQUE constraint prevents duplicate tokens
- For extra paranoia, retry on insert conflict (rare but possible)
