package database

import (
	"database/sql"
	"fmt"
	"log/slog"
)

var migrations = []struct {
	name string
	sql  string
}{
	{
		name: "001_create_users",
		sql: `CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			name TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'viewer',
			is_active BOOLEAN NOT NULL DEFAULT 1,
			must_change_password BOOLEAN NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	},
	{
		name: "002_create_docs",
		sql: `CREATE TABLE IF NOT EXISTS docs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			slug TEXT NOT NULL UNIQUE,
			description TEXT,
			doc_type TEXT NOT NULL DEFAULT 'openapi',
			content TEXT,
			external_url TEXT,
			version TEXT,
			sort_order INTEGER NOT NULL DEFAULT 0,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	},
	{
		name: "003_create_user_doc_access",
		sql: `CREATE TABLE IF NOT EXISTS user_doc_access (
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			doc_id INTEGER NOT NULL REFERENCES docs(id) ON DELETE CASCADE,
			granted_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			granted_by INTEGER REFERENCES users(id),
			PRIMARY KEY (user_id, doc_id)
		);`,
	},
	{
		name: "004_create_migrations_table",
		sql: `CREATE TABLE IF NOT EXISTS _migrations (
			name TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,
	},
	{
		name: "005_create_settings",
		sql: `CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL
		);
		INSERT OR IGNORE INTO settings (key, value) VALUES ('site_title', 'API Docs Portal');`,
	},
}

func runMigrations(db *sql.DB) error {
	// Ensure the migrations tracking table exists first
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
		name TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	for _, m := range migrations {
		if m.name == "004_create_migrations_table" {
			continue
		}

		var exists int
		err := db.QueryRow("SELECT COUNT(*) FROM _migrations WHERE name = ?", m.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", m.name, err)
		}
		if exists > 0 {
			continue
		}

		slog.Info("applying migration", "name", m.name)
		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("apply migration %s: %w", m.name, err)
		}

		if _, err := db.Exec("INSERT INTO _migrations (name) VALUES (?)", m.name); err != nil {
			return fmt.Errorf("record migration %s: %w", m.name, err)
		}
	}

	slog.Info("migrations complete")
	return nil
}
