package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies SQL files from migrationsDir in lexical order.
// It is intentionally simple (no external deps) to keep deploys reliable.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	// Ensure migration table exists.
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		);
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		// If migrations folder is missing in runtime image, this is a hard error.
		return fmt.Errorf("read migrations dir %q: %w", migrationsDir, err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, filepath.Join(migrationsDir, name))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		version := filepath.Base(f)

		var exists bool
		err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if exists {
			continue
		}

		sqlBytes, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}
		sqlText := string(sqlBytes)

		// Apply in a transaction for safety.
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration tx %s: %w", version, err)
		}
		_, err = tx.Exec(ctx, sqlText)
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("exec migration %s: %w", version, err)
		}
		_, err = tx.Exec(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES ($1, $2)`, version, time.Now().UTC())
		if err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}

	return nil
}
