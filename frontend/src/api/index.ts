import { apiFetch, storeToken } from "./client";
import { Achievement, Reward, Task, WorkspaceSnapshot } from "../storage";

export async function register(email: string, password: string) {
  await apiFetch("/auth/register", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
}

export async function login(email: string, password: string) {
  const data = await apiFetch<{ access_token: string }>("/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password })
  });
  storeToken(data.access_token);
}

export async function getMe() {
  return apiFetch<{ id: string; email: string }>("/me");
}

export async function listWorkspaces(): Promise<string | null> {
  const data = await apiFetch<{ workspaces: string[] }>("/workspaces");
  return data.workspaces?.[0] ?? null;
}

export async function fetchWorkspaceBalance(workspaceId: string) {
  return apiFetch<{ workspace_id: string; balance: number }>(`/workspaces/${workspaceId}/balance`);
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
  const data = await apiFetch<{ id: string }>("/tasks", { method: "POST", body: JSON.stringify(payload) });
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

export async function updateTask(workspaceId: string, taskId: string, payload: Partial<Task>) {
  await apiFetch(`/tasks/${taskId}`, {
    method: "PUT",
    body: JSON.stringify({ workspace_id: workspaceId, ...payload })
  });
}

export async function deleteTask(workspaceId: string, taskId: string) {
  await apiFetch(`/tasks/${taskId}?workspace_id=${workspaceId}`, { method: "DELETE" });
}

export async function addReward(
  workspaceId: string,
  form: { title: string; cost: number; description: string; icon: string }
): Promise<Reward> {
  const payload = {
    workspace_id: workspaceId,
    title: form.title,
    description: form.description,
    cost: form.cost,
    is_shared: false
  };
  const data = await apiFetch<{ id: string }>("/rewards", { method: "POST", body: JSON.stringify(payload) });
  return {
    id: data.id,
    workspace_id: workspaceId,
    title: form.title,
    description: form.description,
    cost: form.cost,
    icon: form.icon
  };
}

export async function buyReward(workspaceId: string, rewardId: string) {
  await apiFetch(`/rewards/${rewardId}/buy`, {
    method: "POST",
    body: JSON.stringify({ workspace_id: workspaceId })
  });
}

export async function updateReward(workspaceId: string, rewardId: string, payload: Partial<Reward>) {
  await apiFetch(`/rewards/${rewardId}`, {
    method: "PUT",
    body: JSON.stringify({ workspace_id: workspaceId, ...payload })
  });
}

export async function deleteReward(workspaceId: string, rewardId: string) {
  await apiFetch(`/rewards/${rewardId}?workspace_id=${workspaceId}`, { method: "DELETE" });
}

export async function addAchievement(
  workspaceId: string,
  form: { title: string; description: string; imageUrl: string }
): Promise<Achievement> {
  const payload = {
    workspace_id: workspaceId,
    title: form.title,
    description: form.description,
    image_url: form.imageUrl || null
  };
  const data = await apiFetch<{ id: string }>("/achievements", { method: "POST", body: JSON.stringify(payload) });
  return {
    id: data.id,
    workspace_id: workspaceId,
    title: form.title,
    description: form.description,
    image_url: form.imageUrl || null
  };
}

export async function updateAchievement(
  workspaceId: string,
  achievementId: string,
  payload: Partial<Achievement>
) {
  await apiFetch(`/achievements/${achievementId}`, {
    method: "PUT",
    body: JSON.stringify({ workspace_id: workspaceId, ...payload })
  });
}

export async function deleteAchievement(workspaceId: string, achievementId: string) {
  await apiFetch(`/achievements/${achievementId}?workspace_id=${workspaceId}`, { method: "DELETE" });
}

export async function runSyncPull(snapshot: WorkspaceSnapshot): Promise<WorkspaceSnapshot> {
  if (!snapshot.workspaceId) return snapshot;
  const since = snapshot.lastSync ?? new Date(0).toISOString();
  const data = await apiFetch<{ changes: Record<string, unknown[]>; server_time: string }>(
    `/sync?workspace_id=${snapshot.workspaceId}&since=${since}`
  );
  const changes = data.changes ?? {};
  const serverTime = data.server_time ?? new Date().toISOString();
  const updated: WorkspaceSnapshot = {
    ...snapshot,
    tasks: mergeEntities(snapshot.tasks, (changes.tasks as Task[]) ?? []),
    rewards: mergeEntities(snapshot.rewards, (changes.rewards as Reward[]) ?? []),
    achievements: mergeEntities(snapshot.achievements, (changes.achievements as Achievement[]) ?? []),
    lastSync: serverTime
  };
  return updated;
}

function mergeEntities<T extends { id: string }>(current: T[], incoming: T[]): T[] {
  const map = new Map(current.map((item) => [item.id, item]));
  incoming.forEach((item) => map.set(item.id, item));
  return Array.from(map.values());
}
