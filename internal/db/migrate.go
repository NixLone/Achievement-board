package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies SQL files from migrationsDir in lexical order.
// It is intentionally simple (no external deps) to keep deploys reliable.
//
// Important: if you deploy to a DB that already has tables (e.g. you created them earlier manually),
// this migrator will "baseline" those migrations: it will treat certain "already exists" errors as
// success and mark the migration version as applied.
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
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version=$1)`, version).Scan(&exists); err != nil {
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

			// Baseline: if schema already exists, some CREATE statements may fail with "already exists".
			// In that case we mark the migration as applied and continue.
			if isIgnorableMigrationError(err) {
				if _, recErr := pool.Exec(ctx,
					`INSERT INTO schema_migrations(version, applied_at)
					 VALUES ($1, $2) ON CONFLICT (version) DO NOTHING`,
					version, time.Now().UTC(),
				); recErr != nil {
					return fmt.Errorf("record baseline migration %s: %w", version, recErr)
				}
				continue
			}

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

func isIgnorableMigrationError(err error) bool {
	var pgErr *pgconn.PgError
	if err == nil {
		return false
	}
	if !errors.As(err, &pgErr) {
		return false
	}
	// 42P07 = duplicate_table
	// 42710 = duplicate_object (e.g., extension/type already exists)
	// 23505 = unique_violation (defensive)
	switch pgErr.Code {
	case "42P07", "42710", "23505":
		return true
	default:
		return false
	}
}
