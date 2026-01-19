export type Task = {
  id: string;
  workspace_id: string;
  title: string;
  description: string;
  due_date?: string | null;
  value: number;
  status: string;
  done_at?: string | null;
  deleted_at?: string | null;
  updated_at?: string;
  version?: number;
};

export type Reward = {
  id: string;
  workspace_id: string;
  title: string;
  description: string;
  cost: number;
  deleted_at?: string | null;
  updated_at?: string;
  version?: number;
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

export type WorkspaceSnapshot = {
  user?: { id: string; email: string } | null;
  workspaceId?: string | null;
  tasks: Task[];
  rewards: Reward[];
  achievements: Achievement[];
  lastSync?: string | null;
};

const STORAGE_KEY = "firegoals-snapshot";

export function emptySnapshot(): WorkspaceSnapshot {
  return {
    tasks: [],
    rewards: [],
    achievements: [],
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
