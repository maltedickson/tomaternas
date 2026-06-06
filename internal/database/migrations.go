package database

import (
	"fmt"
	"log"
)

type migration struct {
	version     int
	description string
	statements  []string
}

var migrations = []migration{
	{
		version:     1,
		description: "initial schema",
		statements: []string{
			`CREATE TABLE IF NOT EXISTS users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				username TEXT UNIQUE NOT NULL,
				display_name TEXT UNIQUE NOT NULL,
				password_hash TEXT NOT NULL,
				role TEXT NOT NULL,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
			)`,
			`CREATE TABLE IF NOT EXISTS sessions (
				token TEXT PRIMARY KEY,
				user_id INTEGER NOT NULL,
				created_at DATETIME NOT NULL,
				expires_at DATETIME NOT NULL,
				FOREIGN KEY (user_id) REFERENCES users(id)
			)`,
			`CREATE TABLE IF NOT EXISTS recipes (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				title TEXT NOT NULL,
				description TEXT NOT NULL,
				ingredient_sections TEXT NOT NULL,
				instructions TEXT NOT NULL,
				servings TEXT NOT NULL,
				prep_time_seconds INTEGER NOT NULL,
				cook_time_seconds INTEGER NOT NULL,
				meal_types TEXT NOT NULL,
				dietary_tags TEXT NOT NULL,
				other_tags TEXT NOT NULL,
				owner_id INTEGER NOT NULL,
				created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
				FOREIGN KEY (owner_id) REFERENCES users(id)
			)`,
		},
	},
	{
		version:     2,
		description: "add prep_instructions to recipes",
		statements: []string{
			`ALTER TABLE recipes ADD COLUMN prep_instructions TEXT NOT NULL DEFAULT ''`,
		},
	},
}

func (db *DB) RunMigrations() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("ensuring schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		var count int
		err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`, m.version).Scan(&count)
		if err != nil {
			return fmt.Errorf("checking migration %d, %w", m.version, err)
		}
		if count > 0 {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("starting transaction for migration %d: %w", m.version, err)
		}

		for _, stmt := range m.statements {
			if _, err := tx.Exec(stmt); err != nil {
				tx.Rollback()
				return fmt.Errorf("migration %d (%s): %w", m.version, m.description, err)
			}
		}

		if _, err := tx.Exec(
			`INSERT INTO schema_migrations (version, description) VALUES (?, ?)`,
			m.version, m.description,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("recording migration %d: %w", m.version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("committing migration %d: %w", m.version, err)
		}

		log.Printf("Applied migration %d: %s", m.version, m.description)
	}

	return nil
}
