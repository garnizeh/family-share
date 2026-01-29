# Task 190: Reliability — Health Check DB Ping

**Milestone:** Reliability & Ops  
**Points:** 1 (2 hours)  
**Dependencies:** 052  
**Branch:** `feat/health-db-ping`  
**Labels:** `ops`, `healthcheck`

## Description
Improve `/health` to include a lightweight database ping so failures are detected early.

## Acceptance Criteria
- [ ] Health check returns 200 only if DB is reachable
- [ ] Returns 500 if DB is unavailable
- [ ] Still responds quickly (no heavy queries)

## Files to Add/Modify
- `internal/handler/health.go`
- `internal/handler/health_test.go`

## Tests Required
- [ ] Unit: health returns 200 when DB is up
- [ ] Unit: health returns 500 when DB is down

## PR Checklist
- [ ] Minimal latency impact

## Git Workflow
```bash
git checkout -b feat/health-db-ping
# Update health check endpoint

go test ./internal/handler/... -v

git add internal/handler/

git commit -m "feat: health check validates database connectivity"

git push origin feat/health-db-ping
```
