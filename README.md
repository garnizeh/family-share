# FamilyShare

Lightweight, self-hosted photo sharing for families on low-resource VPS.

This repository contains the backend and SSR frontend for FamilyShare. See `.docs/` for the Technical Design Document and task backlog.

Quick start:

1. Copy `.env.example` to `.env` and fill values.
2. Generate admin password hash:

```bash
make hash-password PASSWORD=YourSecurePassword123
```

Or manually:

```bash
go run scripts/hash_password.go YourSecurePassword123
```

3. Add the generated hash to your environment:

```bash
export ADMIN_PASSWORD_HASH='$2a$12$...'
```

4. Build and run:

```bash
go build -o familyshare ./cmd/app
./familyshare
```

For full setup instructions see the TDD in `.docs/familyshare-tdd.md`.

## Database migrations

Migrations are embedded in the binary under `sql/schema/*` and applied at startup by the DB initializer in `internal/db`.

The helper `InitDB` opens the SQLite database file, enables foreign keys and applies embedded migrations.

Usage example (server bootstrap):

```go
db, err := db.InitDB("./data/familyshare.db")
if err != nil {
	// handle error
}
defer db.Close()
```

If you prefer to apply the migration manually for quick debugging, you can run:

```bash
sqlite3 ./data/familyshare.db < sql/schema/0001_init_schema.sql
```

The migration runner records applied versions in `schema_migrations` so re-running is safe.
