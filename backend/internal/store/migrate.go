package store

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// MigrationsDir resolves the goose migrations directory.
// Prefer MIGRATIONS_DIR, else ./migrations when cwd is backend/.
func MigrationsDir() (string, error) {
	if dir := os.Getenv("MIGRATIONS_DIR"); dir != "" {
		return dir, nil
	}
	if _, err := os.Stat("migrations"); err == nil {
		return filepath.Abs("migrations")
	}
	return "", fmt.Errorf("migrations directory not found: run from backend/ or set MIGRATIONS_DIR")
}

// MigrateUp applies all pending goose migrations against databaseURL.
func MigrateUp(databaseURL string) error {
	dir, err := MigrationsDir()
	if err != nil {
		return err
	}

	cfg, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return fmt.Errorf("parse database url: %w", err)
	}

	db := stdlib.OpenDB(*cfg)
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, dir)
}

// OpenDB opens a database/sql handle using pgx (for tests and future store).
func OpenDB(databaseURL string) (*sql.DB, error) {
	cfg, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	return stdlib.OpenDB(*cfg), nil
}
