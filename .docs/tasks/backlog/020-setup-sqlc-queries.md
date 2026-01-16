# Task 020: Setup sqlc and Generate Query Code

**Milestone:** Setup  
**Points:** 2 (6 hours)  
**Dependencies:** 015  
**Branch:** `feat/sqlc`  
**Labels:** `database`, `codegen`

## Description
Configure sqlc for type-safe SQL queries and generate Go code for all CRUD operations defined in the TDD query catalog.

## Acceptance Criteria
- [ ] `sqlc.yaml` configuration file created
- [ ] SQL query files created in `internal/db/queries/`
- [ ] Generated Go code compiles without errors
- [ ] Repository pattern wrapper created for cleaner API
- [ ] All queries from TDD catalog implemented

## Files to Add/Modify
- `sqlc.yaml` — sqlc configuration
- `sql/queries/albums.sql` — album CRUD queries
- `sql/queries/photos.sql` — photo CRUD queries
- `sql/queries/share_links.sql` — share link queries
- `sql/queries/activity_events.sql` — activity queries
- `internal/db/sqlc/` — generated code (gitignored if using generation step)
- `internal/repository/repository.go` — repository interface wrapper

## sqlc Configuration
```yaml
# sqlc.yaml
version: "2"
sql:
  - schema: "sql/schema/"
    queries: "sql/queries/"
    engine: "sqlite"
    gen:
      go:
        package: "sqlc"
        out: "internal/db/sqlc"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_interface: true
        emit_empty_slices: true
```

## Sample Queries (albums.sql)
```sql
-- name: CreateAlbum :one
INSERT INTO albums (title, description)
VALUES (?, ?)
RETURNING *;

-- name: GetAlbum :one
SELECT * FROM albums WHERE id = ?;

-- name: ListAlbums :many
SELECT * FROM albums
ORDER BY created_at DESC
LIMIT ? OFFSET ?;

-- name: UpdateAlbum :exec
UPDATE albums
SET title = ?, description = ?, cover_photo_id = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: DeleteAlbum :exec
DELETE FROM albums WHERE id = ?;
```

## Tests Required
- [ ] Unit test: generated queries compile
- [ ] Integration test: CreateAlbum + GetAlbum roundtrip
- [ ] Integration test: foreign key cascade (delete album → delete photos)

## PR Checklist
- [ ] `sqlc generate` runs without errors
- [ ] Generated code is gitignored or committed (decide and document)
- [ ] All TDD queries have corresponding SQL definitions
- [ ] Repository wrapper provides clean API
- [ ] Tests pass: `go test ./internal/...`

## Git Workflow
```bash
git checkout -b feat/sqlc
# Install sqlc: go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
sqlc generate
go test ./internal/... -v
git add sqlc.yaml sql/queries/ internal/repository/
git commit -m "feat: setup sqlc and generate type-safe queries"
git push origin feat/sqlc
# Open PR: "Setup sqlc for type-safe database queries"
```

## Notes
- Decide: commit generated code or generate in CI/local builds
- Keep query files organized by entity
- Use prepared statements for all queries (sqlc default)
- Repository layer abstracts sqlc for easier testing/mocking

## Notes on schema location
- The repository stores schema/migration SQL under `sql/schema/` and embeds it via `sql/migrations.go`. Configure `sqlc.yaml` to point `schema:` to `sql/schema/` so generated types align with the embedded schema location.
