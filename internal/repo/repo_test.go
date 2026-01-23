package repo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestRepo(t *testing.T) (*Repo, func()) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}
	ctx := context.Background()
	schema := fmt.Sprintf("test_%d", time.Now().UnixNano())
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		t.Fatalf("parse config: %v", err)
	}
	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", schema))
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	_, err = pool.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	if err != nil {
		pool.Close()
		t.Fatalf("create schema: %v", err)
	}
	if err := createTestTables(ctx, pool); err != nil {
		pool.Close()
		t.Fatalf("create tables: %v", err)
	}
	repo := New(pool)
	return repo, func() {
		_, _ = pool.Exec(ctx, fmt.Sprintf("DROP SCHEMA %s CASCADE", schema))
		pool.Close()
	}
}

func createTestTables(ctx context.Context, pool *pgxpool.Pool) error {
	queries := []string{
		`CREATE TABLE users (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), email text, password_hash text, created_at timestamptz DEFAULT now(), updated_at timestamptz DEFAULT now())`,
		`CREATE TABLE workspaces (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text, type text, created_at timestamptz DEFAULT now(), updated_at timestamptz DEFAULT now())`,
		`CREATE TABLE workspace_members (workspace_id uuid, user_id uuid, role text, permissions jsonb DEFAULT '{}'::jsonb, created_at timestamptz DEFAULT now())`,
		`CREATE TABLE tasks (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), workspace_id uuid, title text, description text DEFAULT '', value numeric(10,2) DEFAULT 0, status text, done_at timestamptz, deleted_at timestamptz, updated_at timestamptz DEFAULT now(), version int DEFAULT 1, is_recurring boolean DEFAULT false, recurrence_weekdays smallint[] NULL, start_date date NULL, end_date date NULL, timezone text NULL)`,
		`CREATE TABLE task_occurrences (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), task_id uuid, occurrence_date date NOT NULL, done boolean DEFAULT false, completed_at timestamptz NULL, created_at timestamptz DEFAULT now())`,
		`CREATE TABLE rewards (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), workspace_id uuid, title text, description text DEFAULT '', cost numeric(10,2), deleted_at timestamptz, updated_at timestamptz DEFAULT now(), version int DEFAULT 1, one_time boolean DEFAULT false)`,
		`CREATE TABLE reward_purchases (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), workspace_id uuid, reward_id uuid, user_id uuid, cost numeric(10,2), purchased_at timestamptz DEFAULT now())`,
		`CREATE TABLE transactions (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), workspace_id uuid, user_id uuid, type text, amount numeric(10,2), reason text, entity_type text, entity_id uuid, created_at timestamptz DEFAULT now())`,
		`CREATE TABLE workspace_balance (workspace_id uuid PRIMARY KEY, balance numeric(10,2) DEFAULT 0, updated_at timestamptz DEFAULT now())`,
		`CREATE TABLE goals (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), workspace_id uuid, title text, description text DEFAULT '', period text DEFAULT 'day', status text DEFAULT 'active', updated_at timestamptz DEFAULT now(), deleted_at timestamptz, version int DEFAULT 1)`,
	}
	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func TestCompleteTaskIdempotent(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	var workspaceID string
	if err := repo.Pool.QueryRow(ctx, `INSERT INTO workspaces (name, type) VALUES ('Test', 'personal') RETURNING id`).Scan(&workspaceID); err != nil {
		t.Fatalf("workspace: %v", err)
	}
	var userID string
	if err := repo.Pool.QueryRow(ctx, `INSERT INTO users (email, password_hash) VALUES ('a@b.com', 'x') RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("user: %v", err)
	}
	if _, err := repo.Pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role) VALUES ($1, $2, 'owner')`, workspaceID, userID); err != nil {
		t.Fatalf("member: %v", err)
	}
	if _, err := repo.Pool.Exec(ctx, `INSERT INTO workspace_balance (workspace_id, balance) VALUES ($1, 0)`, workspaceID); err != nil {
		t.Fatalf("balance: %v", err)
	}
	var taskID string
	if err := repo.Pool.QueryRow(ctx, `INSERT INTO tasks (workspace_id, title, value, status) VALUES ($1, 'Task', 5, 'open') RETURNING id`, workspaceID).Scan(&taskID); err != nil {
		t.Fatalf("task: %v", err)
	}

	value, completed, err := repo.CompleteTask(ctx, taskID, workspaceID, userID, nil)
	if err != nil || !completed || value != 5 {
		t.Fatalf("first complete failed: value=%v completed=%v err=%v", value, completed, err)
	}
	value, completed, err = repo.CompleteTask(ctx, taskID, workspaceID, userID, nil)
	if err != nil || completed || value != 0 {
		t.Fatalf("second complete should be noop: value=%v completed=%v err=%v", value, completed, err)
	}
	var balance float64
	if err := repo.Pool.QueryRow(ctx, `SELECT balance FROM workspace_balance WHERE workspace_id=$1`, workspaceID).Scan(&balance); err != nil {
		t.Fatalf("balance read: %v", err)
	}
	if balance != 5 {
		t.Fatalf("expected balance 5, got %v", balance)
	}
}

func TestBuyRewardInsufficientFunds(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	var workspaceID string
	if err := repo.Pool.QueryRow(ctx, `INSERT INTO workspaces (name, type) VALUES ('Test', 'personal') RETURNING id`).Scan(&workspaceID); err != nil {
		t.Fatalf("workspace: %v", err)
	}
	var userID string
	if err := repo.Pool.QueryRow(ctx, `INSERT INTO users (email, password_hash) VALUES ('c@d.com', 'x') RETURNING id`).Scan(&userID); err != nil {
		t.Fatalf("user: %v", err)
	}
	if _, err := repo.Pool.Exec(ctx, `INSERT INTO workspace_balance (workspace_id, balance) VALUES ($1, 0)`, workspaceID); err != nil {
		t.Fatalf("balance: %v", err)
	}
	var rewardID string
	if err := repo.Pool.QueryRow(ctx, `INSERT INTO rewards (workspace_id, title, cost) VALUES ($1, 'Reward', 10) RETURNING id`, workspaceID).Scan(&rewardID); err != nil {
		t.Fatalf("reward: %v", err)
	}

	if _, err := repo.BuyReward(ctx, rewardID, workspaceID, userID); err == nil || !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("expected insufficient funds, got %v", err)
	}
}

func TestSyncCursorWindow(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()
	ctx := context.Background()

	var workspaceID string
	if err := repo.Pool.QueryRow(ctx, `INSERT INTO workspaces (name, type) VALUES ('Test', 'personal') RETURNING id`).Scan(&workspaceID); err != nil {
		t.Fatalf("workspace: %v", err)
	}
	created := time.Now().UTC().Add(-2 * time.Minute)
	if _, err := repo.Pool.Exec(ctx, `INSERT INTO goals (workspace_id, title, updated_at) VALUES ($1, 'Goal', $2)`, workspaceID, created); err != nil {
		t.Fatalf("goal: %v", err)
	}

	since := time.Now().UTC().Add(-1 * time.Minute)
	until := time.Now().UTC().Add(-30 * time.Second)
	changes, err := repo.ListSyncChanges(ctx, workspaceID, since, until)
	if err != nil {
		t.Fatalf("sync changes: %v", err)
	}
	if len(changes["goals"]) != 0 {
		t.Fatalf("expected no changes in window")
	}

	until = time.Now().UTC().Add(1 * time.Minute)
	changes, err = repo.ListSyncChanges(ctx, workspaceID, since, until)
	if err != nil {
		t.Fatalf("sync changes: %v", err)
	}
	if len(changes["goals"]) != 1 {
		t.Fatalf("expected changes in window")
	}
}
