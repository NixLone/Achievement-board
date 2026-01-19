import { Achievement, Reward, Task, WorkspaceSnapshot } from "./storage";

const API_URL = import.meta.env.VITE_API_URL ?? "http://localhost:8080";

function getToken(): string | null {
  return localStorage.getItem("firegoals-token");
}

function setToken(token: string) {
  localStorage.setItem("firegoals-token", token);
}

async function apiFetch(path: string, options: RequestInit = {}) {
  const token = getToken();
  const headers = new Headers(options.headers ?? {});
  headers.set("Content-Type", "application/json");
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }
  const response = await fetch(`${API_URL}${path}`, { ...options, headers });
  if (!response.ok) {
    throw new Error(`Request failed: ${response.status}`);
  }
  return response.json();
}

export async function register(email: string, password: string) {
  await apiFetch("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function login(email: string, password: string) {
  const data = await apiFetch("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
  setToken(data.access_token);
}

export async function getMe() {
  return apiFetch("/me");
}

export async function listWorkspaces(): Promise<string> {
  const data = await apiFetch("/workspaces");
  return data.workspaces[0];
}

export async function fetchWorkspaceBalance(workspaceId: string) {
  return apiFetch(`/workspaces/${workspaceId}/balance`);
}

export async function addTask(
  workspaceId: string,
  form: { title: string; value: number; dueDate: string }
): Promise<Task> {
  const payload = {
    workspace_id: workspaceId,
    title: form.title,
    value: form.value,
    due_date: form.dueDate || null,
    status: "open",
    description: ""
  };
  const data = await apiFetch("/tasks", { method: "POST", body: JSON.stringify(payload) });
  return {
    id: data.id,
    workspace_id: workspaceId,
    title: form.title,
    description: "",
    due_date: form.dueDate || null,
    value: form.value,
    status: "open"
  };
}

export async function completeTask(workspaceId: string, taskId: string) {
  await apiFetch(`/tasks/${taskId}/complete`, {
    method: "POST",
    body: JSON.stringify({ workspace_id: workspaceId })
  });
}

export async function addReward(
  workspaceId: string,
  form: { title: string; cost: number; description: string }
): Promise<Reward> {
  const payload = {
    workspace_id: workspaceId,
    title: form.title,
    description: form.description,
    cost: form.cost,
    is_shared: false
  };
  const data = await apiFetch("/rewards", { method: "POST", body: JSON.stringify(payload) });
  return {
    id: data.id,
    workspace_id: workspaceId,
    title: form.title,
    description: form.description,
    cost: form.cost
  };
}

export async function buyReward(workspaceId: string, rewardId: string) {
  await apiFetch(`/rewards/${rewardId}/buy`, {
    method: "POST",
    body: JSON.stringify({ workspace_id: workspaceId })
  });
}

export async function addAchievement(
  workspaceId: string,
  form: { title: string; description: string }
): Promise<Achievement> {
  const payload = {
    workspace_id: workspaceId,
    title: form.title,
    description: form.description
  };
  const data = await apiFetch("/achievements", { method: "POST", body: JSON.stringify(payload) });
  return {
    id: data.id,
    workspace_id: workspaceId,
    title: form.title,
    description: form.description
  };
}

export function updateLocalCache(snapshot: WorkspaceSnapshot, updates: Partial<WorkspaceSnapshot>) {
  return {
    ...snapshot,
    ...updates,
    tasks: [...(updates.tasks ?? []), ...snapshot.tasks],
    rewards: [...(updates.rewards ?? []), ...snapshot.rewards],
    achievements: [...(updates.achievements ?? []), ...snapshot.achievements]
  };
}

export async function runSyncPull(snapshot: WorkspaceSnapshot): Promise<WorkspaceSnapshot> {
  if (!snapshot.workspaceId) return snapshot;
  const since = snapshot.lastSync ?? new Date(0).toISOString();
  const data = await apiFetch(`/sync?workspace_id=${snapshot.workspaceId}&since=${since}`);
  const changes = data.changes ?? {};
  const serverTime = data.server_time ?? new Date().toISOString();
  const updated: WorkspaceSnapshot = {
    ...snapshot,
    tasks: mergeEntities(snapshot.tasks, changes.tasks ?? []),
    rewards: mergeEntities(snapshot.rewards, changes.rewards ?? []),
    achievements: mergeEntities(snapshot.achievements, changes.achievements ?? []),
    lastSync: serverTime
  };
  return updated;
}

function mergeEntities<T extends { id: string }>(current: T[], incoming: T[]): T[] {
  const map = new Map(current.map((item) => [item.id, item]));
  incoming.forEach((item) => map.set(item.id, item));
  return Array.from(map.values());
}
