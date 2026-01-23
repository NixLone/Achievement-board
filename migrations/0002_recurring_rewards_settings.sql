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
