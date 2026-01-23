export type Task = {
  id: string;
  workspace_id: string;
  title: string;
  description: string;
  due_date?: string | null;
  is_recurring?: boolean;
  recurrence_weekdays?: number[];
  start_date?: string | null;
  end_date?: string | null;
  timezone?: string | null;
  value: number;
  status: string;
  done_at?: string | null;
  deleted_at?: string | null;
  updated_at?: string;
  version?: number;
};

export type TaskInstance = Task & {
  occurrence_date: string;
  done: boolean;
};

export type Reward = {
  id: string;
  workspace_id: string;
  title: string;
  description: string;
  cost: number;
  icon?: string;
  one_time?: boolean;
  deleted_at?: string | null;
  updated_at?: string;
  version?: number;
};

export type RewardPurchase = {
  id: string;
  workspace_id: string;
  reward_id: string;
  user_id: string;
  cost: number;
  purchased_at: string;
  note?: string | null;
};

export type Achievement = {
  id: string;
  workspace_id: string;
  title: string;
  description: string;
  image_url?: string | null;
  deleted_at?: string | null;
  updated_at?: string;
  version?: number;
};

export type Goal = {
  id: string;
  workspace_id: string;
  title: string;
  description: string;
  period: string;
  status: string;
  updated_at?: string;
  version?: number;
  deleted_at?: string | null;
};

export type WorkspaceSnapshot = {
  user?: { id: string; email: string } | null;
  workspaceId?: string | null;
  settings?: UserSettings | null;
  tasks: Task[];
  rewards: Reward[];
  achievements: Achievement[];
  goals: Goal[];
  lastSync?: string | null;
};

export type UserSettings = {
  user_id: string;
  theme: string;
  last_active_workspace?: string | null;
  updated_at?: string;
};

const STORAGE_KEY = "firegoals-snapshot";

export function emptySnapshot(): WorkspaceSnapshot {
  return {
    tasks: [],
    rewards: [],
    achievements: [],
    goals: [],
    lastSync: null,
    user: null,
    workspaceId: null
  };
}

export async function loadSnapshot(): Promise<WorkspaceSnapshot> {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return emptySnapshot();
  }
  try {
    return JSON.parse(raw) as WorkspaceSnapshot;
  } catch (error) {
    return emptySnapshot();
  }
}

export async function saveSnapshot(snapshot: WorkspaceSnapshot): Promise<void> {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(snapshot));
}
