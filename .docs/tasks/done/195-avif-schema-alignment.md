# Task 195: Storage — AVIF Schema Alignment

**Milestone:** Storage & Compatibility  
**Points:** 1 (3 hours)  
**Dependencies:** 050  
**Branch:** `feat/avif-schema-alignment`  
**Labels:** `storage`, `db`

## Description
Align database schema constraints with actual pipeline output. Either implement AVIF encoding or remove `avif` from `photos.format` constraint.

## Acceptance Criteria
- [ ] Schema and pipeline formats are consistent
- [ ] Migration applied if schema changes
- [ ] Documented decision in `.docs/decisions/`

## Files to Add/Modify
- `sql/schema/0003_update_photo_format.sql` (if removing AVIF)
- `sql/queries/photos.sql` (if needed)
- `.docs/decisions/` — record decision

## Implementation Options
- **Option A:** Add AVIF pipeline support and encode output when enabled.
- **Option B:** Remove AVIF from schema constraint until implemented.

## Tests Required
- [ ] Unit: format constraint aligns with pipeline output

## PR Checklist
- [ ] Migration tested on existing DB

## Git Workflow
```bash
git checkout -b feat/avif-schema-alignment
# Align schema with pipeline output

go test ./... -v

git add sql/ .docs/decisions/

git commit -m "chore: align photo format schema with pipeline"

git push origin feat/avif-schema-alignment
```
