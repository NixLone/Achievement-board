package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrInviteExpired     = errors.New("invite expired")
	ErrInviteUsed        = errors.New("invite used")
)

type Repo struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{Pool: pool}
}

func (r *Repo) CreateUser(ctx context.Context, email, passwordHash string) (string, error) {
	var id string
	err := r.Pool.QueryRow(ctx, `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id`, email, passwordHash).Scan(&id)
	return id, err
}

func (r *Repo) GetUserByEmail(ctx context.Context, email string) (string, string, error) {
	var id, hash string
	err := r.Pool.QueryRow(ctx, `SELECT id, password_hash FROM users WHERE email=$1`, email).Scan(&id, &hash)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return id, hash, err
}

func (r *Repo) GetUserByID(ctx context.Context, userID string) (string, string, error) {
	var id, email string
	err := r.Pool.QueryRow(ctx, `SELECT id, email FROM users WHERE id=$1`, userID).Scan(&id, &email)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return id, email, err
}

func (r *Repo) CreateSession(ctx context.Context, userID, token string, expiresAt time.Time) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO sessions (user_id, token, expires_at) VALUES ($1, $2, $3)`, userID, token, expiresAt)
	return err
}

func (r *Repo) CreateWorkspace(ctx context.Context, name, workspaceType, ownerID string) (string, error) {
	var id string
	if err := r.Pool.QueryRow(ctx, `INSERT INTO workspaces (name, type) VALUES ($1, $2) RETURNING id`, name, workspaceType).Scan(&id); err != nil {
		return "", err
	}
	_, err := r.Pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role, permissions) VALUES ($1, $2, 'owner', '{"see_balance":true,"see_goals":true}'::jsonb)`, id, ownerID)
	if err != nil {
		return "", err
	}
	_, err = r.Pool.Exec(ctx, `INSERT INTO workspace_balance (workspace_id, balance) VALUES ($1, 0)`, id)
	return id, err
}

func (r *Repo) AddWorkspaceMember(ctx context.Context, workspaceID, userID, role string) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role, permissions) VALUES ($1, $2, $3, '{"see_balance":true,"see_goals":true}'::jsonb) ON CONFLICT DO NOTHING`, workspaceID, userID, role)
	return err
}

func (r *Repo) UserInWorkspace(ctx context.Context, userID, workspaceID string) (bool, error) {
	var exists bool
	err := r.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM workspace_members WHERE workspace_id=$1 AND user_id=$2)`, workspaceID, userID).Scan(&exists)
	return exists, err
}

func (r *Repo) GetWorkspaceRole(ctx context.Context, userID, workspaceID string) (string, error) {
	var role string
	err := r.Pool.QueryRow(ctx, `SELECT role FROM workspace_members WHERE workspace_id=$1 AND user_id=$2`, workspaceID, userID).Scan(&role)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return role, err
}

func (r *Repo) ListUserWorkspaces(ctx context.Context, userID string) ([]string, error) {
	rows, err := r.Pool.Query(ctx, `SELECT workspace_id FROM workspace_members WHERE user_id=$1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *Repo) CreateInvite(ctx context.Context, workspaceID, createdBy, code string, expiresAt time.Time) error {
	_, err := r.Pool.Exec(ctx, `INSERT INTO workspace_invites (workspace_id, created_by_user_id, code, expires_at) VALUES ($1, $2, $3, $4)`, workspaceID, createdBy, code, expiresAt)
	return err
}

func (r *Repo) AcceptInvite(ctx context.Context, code, userID string) (string, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	var workspaceID string
	err = tx.QueryRow(ctx, `UPDATE workspace_invites SET used_at=now()
		WHERE code=$1 AND used_at IS NULL AND expires_at > now()
		RETURNING workspace_id`, code).Scan(&workspaceID)
	if errors.Is(err, pgx.ErrNoRows) {
		var expiresAt time.Time
		var usedAt *time.Time
		checkErr := tx.QueryRow(ctx, `SELECT expires_at, used_at FROM workspace_invites WHERE code=$1`, code).Scan(&expiresAt, &usedAt)
		if errors.Is(checkErr, pgx.ErrNoRows) {
			return "", ErrNotFound
		}
		if usedAt != nil {
			return "", ErrInviteUsed
		}
		if time.Now().After(expiresAt) {
			return "", ErrInviteExpired
		}
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO workspace_members (workspace_id, user_id, role, permissions)
		VALUES ($1, $2, 'member', '{"see_balance":true,"see_goals":true}'::jsonb)
		ON CONFLICT DO NOTHING`, workspaceID, userID); err != nil {
		return "", err
	}
	if err := tx.Commit(ctx); err != nil {
		return "", err
	}
	return workspaceID, nil
}

func (r *Repo) CreateGoal(ctx context.Context, workspaceID, title, description, period, status string, startDate, endDate *time.Time) (string, error) {
	var id string
	err := r.Pool.QueryRow(ctx, `INSERT INTO goals (workspace_id, title, description, period, start_date, end_date, status) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`, workspaceID, title, description, period, startDate, endDate, status).Scan(&id)
	return id, err
}

func (r *Repo) UpdateGoal(ctx context.Context, id, workspaceID, title, description, period, status string, startDate, endDate *time.Time) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE goals SET title=$1, description=$2, period=$3, status=$4, start_date=$5, end_date=$6, updated_at=now(), version=version+1 WHERE id=$7 AND workspace_id=$8`, title, description, period, status, startDate, endDate, id, workspaceID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) DeleteGoal(ctx context.Context, id, workspaceID string) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE goals SET deleted_at=now(), updated_at=now(), version=version+1 WHERE id=$1 AND workspace_id=$2`, id, workspaceID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) ListGoals(ctx context.Context, workspaceID string) ([]map[string]any, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, title, description, period, start_date, end_date, status, created_at, updated_at, deleted_at, version FROM goals WHERE workspace_id=$1 AND deleted_at IS NULL`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []map[string]any
	for rows.Next() {
		var id, title, description, period, status string
		var startDate, endDate, deletedAt *time.Time
		var createdAt, updatedAt time.Time
		var version int
		if err := rows.Scan(&id, &title, &description, &period, &startDate, &endDate, &status, &createdAt, &updatedAt, &deletedAt, &version); err != nil {
			return nil, err
		}
		res = append(res, map[string]any{
			"id": id, "workspace_id": workspaceID, "title": title, "description": description, "period": period, "start_date": startDate, "end_date": endDate, "status": status, "created_at": createdAt, "updated_at": updatedAt, "deleted_at": deletedAt, "version": version,
		})
	}
	return res, rows.Err()
}

func (r *Repo) CreateTask(ctx context.Context, workspaceID string, goalID *string, title, description string, dueDate *time.Time, repeatRule *string, value float64, status string) (string, error) {
	var id string
	err := r.Pool.QueryRow(ctx, `INSERT INTO tasks (workspace_id, goal_id, title, description, due_date, repeat_rule, value, status) VALUES ($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`, workspaceID, goalID, title, description, dueDate, repeatRule, value, status).Scan(&id)
	return id, err
}

func (r *Repo) UpdateTask(ctx context.Context, id, workspaceID string, goalID *string, title, description string, dueDate *time.Time, repeatRule *string, value float64, status string) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE tasks SET goal_id=$1, title=$2, description=$3, due_date=$4, repeat_rule=$5, value=$6, status=$7, updated_at=now(), version=version+1 WHERE id=$8 AND workspace_id=$9`, goalID, title, description, dueDate, repeatRule, value, status, id, workspaceID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) CompleteTask(ctx context.Context, id, workspaceID, userID string) (float64, bool, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, false, err
	}
	defer tx.Rollback(ctx)

	var value float64
	err = tx.QueryRow(ctx, `UPDATE tasks SET status='done', done_at=now(), updated_at=now(), version=version+1
		WHERE id=$1 AND workspace_id=$2 AND status!='done' AND deleted_at IS NULL
		RETURNING value`, id, workspaceID).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		var status string
		checkErr := tx.QueryRow(ctx, `SELECT status FROM tasks WHERE id=$1 AND workspace_id=$2 AND deleted_at IS NULL`, id, workspaceID).Scan(&status)
		if errors.Is(checkErr, pgx.ErrNoRows) {
			return 0, false, ErrNotFound
		}
		if status == "done" {
			return 0, false, nil
		}
		return 0, false, checkErr
	}
	if err != nil {
		return 0, false, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO transactions (workspace_id, user_id, type, amount, reason, entity_type, entity_id)
		VALUES ($1,$2,'earn',$3,'task completed','task',$4)`, workspaceID, userID, value, id); err != nil {
		return 0, false, err
	}
	if _, err := tx.Exec(ctx, `UPDATE workspace_balance SET balance = balance + $1, updated_at=now() WHERE workspace_id=$2`, value, workspaceID); err != nil {
		return 0, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, false, err
	}
	return value, true, nil
}

func (r *Repo) DeleteTask(ctx context.Context, id, workspaceID string) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE tasks SET deleted_at=now(), updated_at=now(), version=version+1 WHERE id=$1 AND workspace_id=$2`, id, workspaceID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) ListTasks(ctx context.Context, workspaceID string) ([]map[string]any, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, goal_id, title, description, due_date, repeat_rule, value, status, done_at, created_at, updated_at, deleted_at, version FROM tasks WHERE workspace_id=$1 AND deleted_at IS NULL`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []map[string]any
	for rows.Next() {
		var id string
		var goalID *string
		var title, description, status string
		var dueDate, doneAt, deletedAt *time.Time
		var repeatRule *string
		var value float64
		var createdAt, updatedAt time.Time
		var version int
		if err := rows.Scan(&id, &goalID, &title, &description, &dueDate, &repeatRule, &value, &status, &doneAt, &createdAt, &updatedAt, &deletedAt, &version); err != nil {
			return nil, err
		}
		res = append(res, map[string]any{
			"id": id, "workspace_id": workspaceID, "goal_id": goalID, "title": title, "description": description, "due_date": dueDate, "repeat_rule": repeatRule, "value": value, "status": status, "done_at": doneAt, "created_at": createdAt, "updated_at": updatedAt, "deleted_at": deletedAt, "version": version,
		})
	}
	return res, rows.Err()
}

func (r *Repo) CreateReward(ctx context.Context, workspaceID, title, description string, cost float64, isShared bool, cooldownHours *int) (string, error) {
	var id string
	err := r.Pool.QueryRow(ctx, `INSERT INTO rewards (workspace_id, title, description, cost, is_shared, cooldown_hours) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`, workspaceID, title, description, cost, isShared, cooldownHours).Scan(&id)
	return id, err
}

func (r *Repo) UpdateReward(ctx context.Context, id, workspaceID, title, description string, cost float64, isShared bool, cooldownHours *int) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE rewards SET title=$1, description=$2, cost=$3, is_shared=$4, cooldown_hours=$5, updated_at=now(), version=version+1 WHERE id=$6 AND workspace_id=$7`, title, description, cost, isShared, cooldownHours, id, workspaceID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) DeleteReward(ctx context.Context, id, workspaceID string) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE rewards SET deleted_at=now(), updated_at=now(), version=version+1 WHERE id=$1 AND workspace_id=$2`, id, workspaceID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) ListRewards(ctx context.Context, workspaceID string) ([]map[string]any, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, title, description, cost, is_shared, cooldown_hours, created_at, updated_at, deleted_at, version FROM rewards WHERE workspace_id=$1 AND deleted_at IS NULL`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []map[string]any
	for rows.Next() {
		var id, title, description string
		var cost float64
		var isShared bool
		var cooldownHours *int
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time
		var version int
		if err := rows.Scan(&id, &title, &description, &cost, &isShared, &cooldownHours, &createdAt, &updatedAt, &deletedAt, &version); err != nil {
			return nil, err
		}
		res = append(res, map[string]any{
			"id": id, "workspace_id": workspaceID, "title": title, "description": description, "cost": cost, "is_shared": isShared, "cooldown_hours": cooldownHours, "created_at": createdAt, "updated_at": updatedAt, "deleted_at": deletedAt, "version": version,
		})
	}
	return res, rows.Err()
}

func (r *Repo) BuyReward(ctx context.Context, rewardID, workspaceID, userID string) (float64, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var cost float64
	if err := tx.QueryRow(ctx, `SELECT cost FROM rewards WHERE id=$1 AND workspace_id=$2 AND deleted_at IS NULL`, rewardID, workspaceID).Scan(&cost); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}
	var balance float64
	if err := tx.QueryRow(ctx, `UPDATE workspace_balance SET balance = balance - $1, updated_at=now()
		WHERE workspace_id=$2 AND balance >= $1 RETURNING balance`, cost, workspaceID).Scan(&balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, ErrInsufficientFunds
		}
		return 0, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reward_purchases (workspace_id, reward_id, user_id, cost)
		VALUES ($1,$2,$3,$4)`, workspaceID, rewardID, userID, cost); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO transactions (workspace_id, user_id, type, amount, reason, entity_type, entity_id)
		VALUES ($1,$2,'spend',$3,'reward purchased','reward',$4)`, workspaceID, userID, cost, rewardID); err != nil {
		return 0, err
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return cost, nil
}

func (r *Repo) CreateAchievement(ctx context.Context, workspaceID, title, description string, imageURL *string) (string, error) {
	var id string
	err := r.Pool.QueryRow(ctx, `INSERT INTO achievements (workspace_id, title, description, image_url) VALUES ($1,$2,$3,$4) RETURNING id`, workspaceID, title, description, imageURL).Scan(&id)
	return id, err
}

func (r *Repo) UpdateAchievement(ctx context.Context, id, workspaceID, title, description string, imageURL *string, achievedAt *time.Time) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE achievements SET title=$1, description=$2, image_url=$3, achieved_at=$4, updated_at=now(), version=version+1 WHERE id=$5 AND workspace_id=$6`, title, description, imageURL, achievedAt, id, workspaceID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) DeleteAchievement(ctx context.Context, id, workspaceID string) error {
	cmd, err := r.Pool.Exec(ctx, `UPDATE achievements SET deleted_at=now(), updated_at=now(), version=version+1 WHERE id=$1 AND workspace_id=$2`, id, workspaceID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) ListAchievements(ctx context.Context, workspaceID string) ([]map[string]any, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, title, description, image_url, achieved_at, created_at, updated_at, deleted_at, version FROM achievements WHERE workspace_id=$1 AND deleted_at IS NULL`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var res []map[string]any
	for rows.Next() {
		var id, title, description string
		var imageURL *string
		var achievedAt, deletedAt *time.Time
		var createdAt, updatedAt time.Time
		var version int
		if err := rows.Scan(&id, &title, &description, &imageURL, &achievedAt, &createdAt, &updatedAt, &deletedAt, &version); err != nil {
			return nil, err
		}
		res = append(res, map[string]any{
			"id": id, "workspace_id": workspaceID, "title": title, "description": description, "image_url": imageURL, "achieved_at": achievedAt, "created_at": createdAt, "updated_at": updatedAt, "deleted_at": deletedAt, "version": version,
		})
	}
	return res, rows.Err()
}

func (r *Repo) ListSyncChanges(ctx context.Context, workspaceID string, since, until time.Time) (map[string][]map[string]any, error) {
	goals, err := r.queryEntity(ctx, `SELECT id, title, description, period, start_date, end_date, status, created_at, updated_at, deleted_at, version
		FROM goals WHERE workspace_id=$1 AND ((updated_at > $2 AND updated_at <= $3) OR (deleted_at IS NOT NULL AND deleted_at > $2 AND deleted_at <= $3))`, workspaceID, since, until)
	if err != nil {
		return nil, err
	}
	tasks, err := r.queryEntity(ctx, `SELECT id, goal_id, title, description, due_date, repeat_rule, value, status, done_at, created_at, updated_at, deleted_at, version
		FROM tasks WHERE workspace_id=$1 AND ((updated_at > $2 AND updated_at <= $3) OR (deleted_at IS NOT NULL AND deleted_at > $2 AND deleted_at <= $3))`, workspaceID, since, until)
	if err != nil {
		return nil, err
	}
	rewards, err := r.queryEntity(ctx, `SELECT id, title, description, cost, is_shared, cooldown_hours, created_at, updated_at, deleted_at, version
		FROM rewards WHERE workspace_id=$1 AND ((updated_at > $2 AND updated_at <= $3) OR (deleted_at IS NOT NULL AND deleted_at > $2 AND deleted_at <= $3))`, workspaceID, since, until)
	if err != nil {
		return nil, err
	}
	achievements, err := r.queryEntity(ctx, `SELECT id, title, description, image_url, achieved_at, created_at, updated_at, deleted_at, version
		FROM achievements WHERE workspace_id=$1 AND ((updated_at > $2 AND updated_at <= $3) OR (deleted_at IS NOT NULL AND deleted_at > $2 AND deleted_at <= $3))`, workspaceID, since, until)
	if err != nil {
		return nil, err
	}
	return map[string][]map[string]any{
		"goals":        goals,
		"tasks":        tasks,
		"rewards":      rewards,
		"achievements": achievements,
	}, nil
}

func (r *Repo) queryEntity(ctx context.Context, query string, args ...any) ([]map[string]any, error) {
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	fields := rows.FieldDescriptions()
	var results []map[string]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		entry := make(map[string]any)
		for i, field := range fields {
			entry[string(field.Name)] = values[i]
		}
		entry["workspace_id"] = args[0]
		results = append(results, entry)
	}
	return results, rows.Err()
}

func (r *Repo) UpsertEntity(ctx context.Context, table string, payload map[string]any) error {
	switch table {
	case "goals":
		_, err := r.Pool.Exec(ctx, `INSERT INTO goals (id, workspace_id, title, description, period, start_date, end_date, status, created_at, updated_at, deleted_at, version)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE($9, now()),COALESCE($10, now()),$11,COALESCE($12, 1))
			ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description, period=EXCLUDED.period, start_date=EXCLUDED.start_date, end_date=EXCLUDED.end_date, status=EXCLUDED.status, updated_at=EXCLUDED.updated_at, deleted_at=EXCLUDED.deleted_at, version=EXCLUDED.version
			WHERE goals.updated_at < EXCLUDED.updated_at OR (goals.updated_at = EXCLUDED.updated_at AND goals.version < EXCLUDED.version)`,
			payload["id"], payload["workspace_id"], payload["title"], payload["description"], payload["period"], payload["start_date"], payload["end_date"], payload["status"], payload["created_at"], payload["updated_at"], payload["deleted_at"], payload["version"])
		return err
	case "tasks":
		_, err := r.Pool.Exec(ctx, `INSERT INTO tasks (id, workspace_id, goal_id, title, description, due_date, repeat_rule, value, status, done_at, created_at, updated_at, deleted_at, version)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,COALESCE($11, now()),COALESCE($12, now()),$13,COALESCE($14, 1))
			ON CONFLICT (id) DO UPDATE SET goal_id=EXCLUDED.goal_id, title=EXCLUDED.title, description=EXCLUDED.description, due_date=EXCLUDED.due_date, repeat_rule=EXCLUDED.repeat_rule, value=EXCLUDED.value, status=EXCLUDED.status, done_at=EXCLUDED.done_at, updated_at=EXCLUDED.updated_at, deleted_at=EXCLUDED.deleted_at, version=EXCLUDED.version
			WHERE tasks.updated_at < EXCLUDED.updated_at OR (tasks.updated_at = EXCLUDED.updated_at AND tasks.version < EXCLUDED.version)`,
			payload["id"], payload["workspace_id"], payload["goal_id"], payload["title"], payload["description"], payload["due_date"], payload["repeat_rule"], payload["value"], payload["status"], payload["done_at"], payload["created_at"], payload["updated_at"], payload["deleted_at"], payload["version"])
		return err
	case "rewards":
		_, err := r.Pool.Exec(ctx, `INSERT INTO rewards (id, workspace_id, title, description, cost, is_shared, cooldown_hours, created_at, updated_at, deleted_at, version)
			VALUES ($1,$2,$3,$4,$5,$6,$7,COALESCE($8, now()),COALESCE($9, now()),$10,COALESCE($11, 1))
			ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description, cost=EXCLUDED.cost, is_shared=EXCLUDED.is_shared, cooldown_hours=EXCLUDED.cooldown_hours, updated_at=EXCLUDED.updated_at, deleted_at=EXCLUDED.deleted_at, version=EXCLUDED.version
			WHERE rewards.updated_at < EXCLUDED.updated_at OR (rewards.updated_at = EXCLUDED.updated_at AND rewards.version < EXCLUDED.version)`,
			payload["id"], payload["workspace_id"], payload["title"], payload["description"], payload["cost"], payload["is_shared"], payload["cooldown_hours"], payload["created_at"], payload["updated_at"], payload["deleted_at"], payload["version"])
		return err
	case "achievements":
		_, err := r.Pool.Exec(ctx, `INSERT INTO achievements (id, workspace_id, title, description, image_url, achieved_at, created_at, updated_at, deleted_at, version)
			VALUES ($1,$2,$3,$4,$5,$6,COALESCE($7, now()),COALESCE($8, now()),$9,COALESCE($10, 1))
			ON CONFLICT (id) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description, image_url=EXCLUDED.image_url, achieved_at=EXCLUDED.achieved_at, updated_at=EXCLUDED.updated_at, deleted_at=EXCLUDED.deleted_at, version=EXCLUDED.version
			WHERE achievements.updated_at < EXCLUDED.updated_at OR (achievements.updated_at = EXCLUDED.updated_at AND achievements.version < EXCLUDED.version)`,
			payload["id"], payload["workspace_id"], payload["title"], payload["description"], payload["image_url"], payload["achieved_at"], payload["created_at"], payload["updated_at"], payload["deleted_at"], payload["version"])
		return err
	default:
		return errors.New("unknown table")
	}
}

func (r *Repo) GetWorkspaceBalance(ctx context.Context, workspaceID string) (float64, error) {
	var balance float64
	err := r.Pool.QueryRow(ctx, `SELECT balance FROM workspace_balance WHERE workspace_id=$1`, workspaceID).Scan(&balance)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrNotFound
	}
	return balance, err
}
