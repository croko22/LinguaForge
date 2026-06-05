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

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{1, migration001},
	{2, migration002},
	{3, migration003},
	{4, migration004},
}

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range migrations {
		var applied int
		err := db.QueryRow("SELECT version FROM schema_migrations WHERE version = ?", m.version).Scan(&applied)
		if err == sql.ErrNoRows {
			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("migration %03d begin tx: %w", m.version, err)
			}

			if _, err := tx.Exec(m.sql); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %03d: %w", m.version, err)
			}

			if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.version); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %03d record: %w", m.version, err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("migration %03d commit: %w", m.version, err)
			}
		} else if err != nil {
			return fmt.Errorf("migration %03d check: %w", m.version, err)
		}
	}

	return nil
}
