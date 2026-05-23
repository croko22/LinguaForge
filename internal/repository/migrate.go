package repository

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed 001_initial.sql
var migration001 string

// RunMigrations executes all pending database migrations.
// Idempotent — safe to call multiple times (all statements use IF NOT EXISTS).
func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(migration001)
	if err != nil {
		return fmt.Errorf("migration 001: %w", err)
	}
	return nil
}
