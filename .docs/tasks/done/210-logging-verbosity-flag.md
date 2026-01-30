# Task 210: Ops — Configurable Logging Verbosity

**Milestone:** Ops & Observability  
**Points:** 1 (2 hours)  
**Dependencies:** 052  
**Branch:** `feat/logging-verbosity`  
**Labels:** `ops`, `logging`

## Description
Introduce a debug flag to reduce noisy logs (e.g., template names) in production.

## Acceptance Criteria
- [ ] Debug flag configured via env
- [ ] Verbose logs only when debug enabled
- [ ] Default is safe for production (minimal logs)

## Files to Add/Modify
- `internal/config/config.go`
- `internal/handler/handler.go`
- `.env.example`

## Tests Required
- [ ] Unit: debug flag controls template logging

## PR Checklist
- [ ] Docs updated with new env var

## Git Workflow
```bash
git checkout -b feat/logging-verbosity
# Add debug flag for logging

go test ./internal/handler/... -v

git add internal/ .env.example

git commit -m "chore: add debug flag for verbose logging"

git push origin feat/logging-verbosity
```
