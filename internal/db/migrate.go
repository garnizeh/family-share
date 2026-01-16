package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ApplyMigrations applies SQL migrations from migrationsFS to the given db.
func ApplyMigrations(db *sql.DB, migrationsFS embed.FS) error {
	// Ensure schema_migrations exists
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
        version INTEGER PRIMARY KEY,
        applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	// Read applied versions
	rows, err := db.Query(`SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("query applied migrations: %w", err)
	}
	defer rows.Close()
	applied := map[int]bool{}
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return err
		}
		applied[v] = true
	}

	// Read migration files from the embedded FS
	entries, err := fs.ReadDir(migrationsFS, "schema")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	type mig struct {
		name string
		ver  int
	}
	var items []mig
	re := regexp.MustCompile(`^(\d+)`) // leading digits

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		m := re.FindStringSubmatch(name)
		if len(m) < 2 {
			continue
		}
		v, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		items = append(items, mig{name: name, ver: v})
	}

	// Sort migrations by numeric version
	sort.Slice(items, func(i, j int) bool { return items[i].ver < items[j].ver })

	for _, it := range items {
		if applied[it.ver] {
			continue
		}

		path := filepath.Join("schema", it.name)
		b, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s from embed FS: %w", it.name, err)
		}

		// Apply migration in transaction
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin tx: %w", err)
		}
		if _, err := tx.Exec(string(b)); err != nil {
			tx.Rollback()
			return fmt.Errorf("exec migration %s: %w", it.name, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations(version) VALUES(?)`, it.ver); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", it.name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", it.name, err)
		}
	}

	return nil
}
