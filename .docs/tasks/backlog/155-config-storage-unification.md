# Task 155: Config — Unify Storage Base Path

**Milestone:** Config & Ops  
**Points:** 1 (3 hours)  
**Dependencies:** 050  
**Branch:** `feat/storage-config-unify`  
**Labels:** `config`, `storage`

## Description
Ensure the storage base path is sourced from a single configuration value and passed consistently to the pipeline and handlers.

## Acceptance Criteria
- [ ] Pipeline does not read `STORAGE_PATH` directly from env
- [ ] Storage base dir comes from `config.Config`
- [ ] All path functions use the same base dir

## Files to Add/Modify
- `internal/config/config.go` — ensure `DataDir`/`StoragePath` is defined
- `internal/pipeline/save.go` — accept base dir as parameter
- `internal/handler/admin_upload.go` — pass base dir into pipeline
- `internal/storage/storage.go` — align naming (optional)

## Implementation Notes
- Add a `StoragePath` field in config if not already present.
- Pass the base dir into `ProcessAndSave` or `SaveProcessedImage`.

## Tests Required
- [ ] Unit: pipeline uses provided base dir
- [ ] Integration: uploaded photo stored under configured path

## PR Checklist
- [ ] `.env` and `.env.example` documented
- [ ] No direct reads of `os.Getenv("STORAGE_PATH")` in pipeline

## Git Workflow
```bash
git checkout -b feat/storage-config-unify
# Unify storage configuration usage

go test ./internal/pipeline/... -v

go test ./internal/handler/... -v

git add internal/
git commit -m "chore: unify storage base path configuration"
git push origin feat/storage-config-unify
```
