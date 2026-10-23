package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// DB is the global database handle used by the store package.
var DB *sql.DB

// ConnectDB opens the sqlite database at dbPath and verifies the connection.
// The connection is stored in the global DB variable.
func ConnectDB(dbPath string) error {
	// Create the database file with owner-only permissions before SQLite opens
	// it, so API keys are never readable by other users on the system.
	if !strings.HasPrefix(dbPath, ":memory:") && !strings.HasPrefix(dbPath, "file:") {
		file, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0600)

		if err != nil {
			return fmt.Errorf("create database file: %w", err)
		}

		file.Close()

		if err := os.Chmod(dbPath, 0600); err != nil {
			return fmt.Errorf("set database permissions: %w", err)
		}
	}

	db, err := sql.Open("sqlite", dbPath)

	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}

	// Enable WAL mode and foreign keys for better concurrency and integrity.
	// secure_delete overwrites deleted content so stale ciphertext does not
	// linger in free pages.
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA foreign_keys = ON;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA secure_delete = ON;",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return fmt.Errorf("exec pragma %q: %w", pragma, err)
		}
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("ping database: %w", err)
	}

	DB = db

	return nil
}

// CloseDB closes the global database connection and resets it to nil.
func CloseDB() error {
	if DB == nil {
		return nil
	}

	if err := DB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	DB = nil

	return nil
}

// CreateTables creates the tables required by Avenor if they do not exist yet.
func CreateTables() error {
	if DB == nil {
		return errors.New("database not connected")
	}

	// API keys live in a separate table so listing providers never reads them
	// by accident. Only the encrypted form and the nonce are stored.
	const schema = `
CREATE TABLE IF NOT EXISTS providers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	base_url TEXT NOT NULL UNIQUE,
	api_key_last4 TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS provider_secrets (
	provider_id INTEGER PRIMARY KEY REFERENCES providers(id) ON DELETE CASCADE,
	ciphertext BLOB NOT NULL,
	nonce BLOB NOT NULL,
	key_version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS projects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	root_path TEXT NOT NULL UNIQUE,
	created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
`

	if _, err := DB.Exec(schema); err != nil {
		return fmt.Errorf("create tables: %w", err)
	}

	return nil
}
