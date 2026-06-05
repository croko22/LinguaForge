package repository

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed 001_initial.sql
var migration001 string

//go:embed 002_words.sql
var migration002 string

//go:embed 003_covers.sql
var migration003 string

//go:embed 004_srs.sql
var migration004 string

// RunMigrations executes all pending database migrations.
// Idempotent — safe to call multiple times (all statements use IF NOT EXISTS).
func RunMigrations(db *sql.DB) error {
	if _, err := db.Exec(migration001); err != nil {
		return fmt.Errorf("migration 001: %w", err)
	}
	if _, err := db.Exec(migration002); err != nil {
		return fmt.Errorf("migration 002: %w", err)
	}
	if _, err := db.Exec(migration003); err != nil {
		return fmt.Errorf("migration 003: %w", err)
	}
	if _, err := db.Exec(migration004); err != nil {
		return fmt.Errorf("migration 004: %w", err)
	}
	return nil
}
