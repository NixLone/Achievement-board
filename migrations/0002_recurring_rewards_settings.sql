-- Defensive migration: supports existing databases created before recurrence features.
-- Adds recurrence fields and settings, and ensures task_occurrences table exists.

-- Ensure task_occurrences exists (older DBs may not have it).
CREATE TABLE IF NOT EXISTS task_occurrences (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  task_id uuid REFERENCES tasks(id) ON DELETE CASCADE,
  occurrence_date date NOT NULL,
  done boolean NOT NULL DEFAULT false,
  completed_at timestamptz NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  UNIQUE (task_id, occurrence_date)
);

ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS recurrence_weekdays smallint[] NULL;

ALTER TABLE rewards
  ADD COLUMN IF NOT EXISTS one_time boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS user_settings (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  theme text NOT NULL DEFAULT 'light-minimal',
  last_active_workspace uuid REFERENCES workspaces(id) ON DELETE SET NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_task_occurrences_date ON task_occurrences (occurrence_date);
