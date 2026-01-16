# Task 140: Final Testing and Bug Fixes

**Milestone:** Tests & CI  
**Points:** 2 (8 hours)  
**Dependencies:** 135  
**Branch:** `feat/final-testing`  
**Labels:** `testing`, `bugfix`, `quality`

## Description
Comprehensive end-to-end testing of the full MVP, bug fixing, and performance validation before initial release.

## Acceptance Criteria
- [ ] All happy paths tested manually (admin workflow, visitor workflow)
- [ ] All error paths tested (expired links, invalid uploads, etc.)
- [ ] Performance tested: upload 50 photos, check load times
- [ ] Memory leaks checked (upload 100+ photos without crash)
- [ ] All known bugs fixed
- [ ] Go test coverage > 60%
- [ ] No critical security issues

## Testing Scenarios

### Admin Workflow
1. Login with correct/incorrect password
2. Create album with/without description
3. Upload single photo (JPEG, PNG)
4. Upload batch (10 photos)
5. Delete photo, verify file removed
6. Delete album, verify cascade
7. Create share link with max views
8. Create share link with expiration
9. Revoke share link
10. View dashboard metrics

### Visitor Workflow
1. Access valid share link
2. Access expired share link (expect error)
3. Access share link at view limit (expect error)
4. Navigate photo gallery with HTMX
5. Open lightbox, use arrow keys
6. Reload page (should not increment view count)

### Performance Testing
- Upload 50 photos (batch), measure time
- Load album with 100 photos, check page load time
- Share link access under load (simulate 10 concurrent visitors)

### Security Testing
- Attempt admin access without auth (should redirect)
- Brute-force share link tokens (rate limit should block)
- CSRF protection (POST without token should fail)

## Known Issues to Fix
- (List will be populated during testing)

## Tools
- Manual testing in Chrome, Firefox, Safari (mobile + desktop)
- Lighthouse for performance and accessibility
- `go test -race -cover ./...` for unit/integration tests
- `hey` or `ab` for basic load testing

## Tests Required
- [ ] All unit tests pass
- [ ] All integration tests pass
- [ ] Manual E2E tests pass
- [ ] No race conditions detected
- [ ] Performance acceptable on low-end VPS

## PR Checklist
- [ ] All critical bugs fixed
- [ ] Test coverage meets target (> 60%)
- [ ] No known security issues
- [ ] Performance validated
- [ ] Manual testing checklist completed

## Git Workflow
```bash
git checkout -b feat/final-testing
# Run full test suite
go test ./... -v -race -cover
# Manual testing
# Fix bugs
git add .
git commit -m "test: comprehensive E2E testing and bug fixes for MVP"
git push origin feat/final-testing
# Open PR: "Final MVP testing and bug fixes"
```

## Notes
- Create GitHub issues for non-critical bugs (defer to post-MVP)
- Document any workarounds or known limitations
- Performance baseline: < 2s page load, < 5s for 10 photo upload
- Test on actual VPS (not just localhost)
