package repository

import (
	"context"
	"database/sql"

	"familyshare/internal/db/sqlc"
)

// Repository is a minimal interface wrapper around data access.
type Repository struct {
	db *sql.DB
	q  sqlc.Querier
}

// NewWithQuerier creates a repository from an existing sqlc Querier (for testing).
func NewWithQuerier(q sqlc.Querier, db *sql.DB) *Repository {
	return &Repository{q: q, db: db}
}

// New creates a repository by instantiating the generated sqlc Queries.
func New(db *sql.DB) *Repository {
	return NewWithQuerier(sqlc.New(db), db)
}

// Ping checks DB connectivity.
func (r *Repository) Ping(ctx context.Context) error {
	if r.db == nil {
		return nil
	}
	return r.db.PingContext(ctx)
}
