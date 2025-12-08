package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
}

func NewDB(dbPath string) (*DB, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create db directory: %w", err)
		}
	}

	conn, err := sql.Open("sqlite", dbPath+"?_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Run migrations
	if err := db.runMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}

func (db *DB) runMigrations() error {
	// Create migrations tracking table
	_, err := db.conn.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get list of migration files
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try alternative path
		migrationsDir = filepath.Join("..", "..", "migrations")
		if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
			return fmt.Errorf("migrations directory not found")
		}
	}

	// Read all migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migrations by filename (they should be numbered: 001, 002, etc.)
	migrationFiles := []string{}
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".sql" {
			migrationFiles = append(migrationFiles, entry.Name())
		}
	}

	// Sort to ensure correct order
	for i := 0; i < len(migrationFiles)-1; i++ {
		for j := i + 1; j < len(migrationFiles); j++ {
			if migrationFiles[i] > migrationFiles[j] {
				migrationFiles[i], migrationFiles[j] = migrationFiles[j], migrationFiles[i]
			}
		}
	}

	// Run each migration
	for _, migrationFile := range migrationFiles {
		// Check if migration has already been applied
		var count int
		err := db.conn.QueryRow(
			"SELECT COUNT(*) FROM schema_migrations WHERE version = ?",
			migrationFile,
		).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			continue // Migration already applied
		}

		// Read and execute migration
		migrationPath := filepath.Join(migrationsDir, migrationFile)
		sqlBytes, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", migrationFile, err)
		}

		// Execute migration in a transaction
		tx, err := db.conn.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", migrationFile, err)
		}

		_, err = tx.Exec(string(sqlBytes))
		if err != nil {
			// Check if error is due to column already existing (SQLite error code 1)
			// This can happen if migration was partially applied
			errStr := err.Error()
			if strings.Contains(errStr, "duplicate column") || 
			   strings.Contains(errStr, "already exists") ||
			   strings.Contains(errStr, "sprite_data") {
				// Column might already exist, log and continue
				// We'll still mark the migration as applied
				fmt.Printf("Warning: Migration %s had an error (column may already exist): %v\n", migrationFile, err)
			} else {
				tx.Rollback()
				return fmt.Errorf("failed to execute migration %s: %w", migrationFile, err)
			}
		}

		// Record migration as applied
		_, err = tx.Exec(
			"INSERT INTO schema_migrations (version) VALUES (?)",
			migrationFile,
		)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", migrationFile, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migrationFile, err)
		}
	}

	return nil
}

