# Task 015: Create Initial SQL Migrations

**Milestone:** Setup  
**Points:** 1 (4 hours)  
**Dependencies:** 010  
**Branch:** `feat/migrations`  
**Labels:** `database`, `setup`

## Description
Create SQL migration files for the initial database schema including albums, photos, share_links, share_link_views, and activity_events tables.

## Acceptance Criteria
- [ ] Migration file `0001_init_schema.sql` created with all core tables
- [ ] Indexes defined for performance (token lookup, foreign keys, created_at)
- [ ] Foreign key constraints enabled
- [ ] Migration guard table (`schema_migrations`) created
- [ ] Simple migration runner implemented in `internal/db/migrate.go`

## Files to Add/Modify
- `migrations/0001_init_schema.sql` — initial schema DDL
- `internal/db/migrate.go` — migration runner logic
- `internal/db/db.go` — database connection and initialization

## Migration File Structure
```sql
-- migrations/0001_init_schema.sql
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE albums (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    cover_photo_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE photos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    album_id INTEGER NOT NULL,
    filename TEXT NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    size_bytes INTEGER NOT NULL,
    format TEXT NOT NULL CHECK(format IN ('webp', 'avif')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (album_id) REFERENCES albums(id) ON DELETE CASCADE
);

CREATE TABLE share_links (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT UNIQUE NOT NULL,
    target_type TEXT NOT NULL CHECK(target_type IN ('album', 'photo')),
    target_id INTEGER NOT NULL,
    max_views INTEGER,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    revoked_at DATETIME
);

CREATE TABLE share_link_views (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    share_link_id INTEGER NOT NULL,
    viewer_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (share_link_id) REFERENCES share_links(id) ON DELETE CASCADE
);

CREATE TABLE activity_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    event_type TEXT NOT NULL,
    album_id INTEGER,
    photo_id INTEGER,
    share_link_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_photos_album_id ON photos(album_id);
CREATE UNIQUE INDEX idx_share_links_token ON share_links(token);
CREATE UNIQUE INDEX idx_share_link_views_dedup ON share_link_views(share_link_id, viewer_hash);
CREATE INDEX idx_activity_events_created_at ON activity_events(created_at);
```

## Tests Required
- [ ] Unit test: migration runs successfully on empty DB
- [ ] Unit test: re-running migration is idempotent
- [ ] Unit test: foreign key constraints work (cascade delete)

## PR Checklist
- [ ] Migration file syntax validated with `sqlite3 :memory: < migrations/0001_init_schema.sql`
- [ ] All tables have primary keys
- [ ] All foreign keys have indexes
- [ ] Migration runner code reviewed
- [ ] Tests pass: `go test ./internal/db/...`

## Git Workflow
```bash
git checkout -b feat/migrations
# Create migration file and runner
go test ./internal/db/... -v
git add migrations/ internal/db/
git commit -m "feat: create initial database schema migration"
git push origin feat/migrations
# Open PR: "Add initial SQL schema migration"
```

## Notes
- Keep migrations simple and forward-only for MVP
- Document migration naming convention: `NNNN_description.sql`
- Migration runner should check version before applying
- Use transactions for migration safety
