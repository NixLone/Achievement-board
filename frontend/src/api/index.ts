import { apiFetch, storeToken } from "./client";
import { Achievement, Reward, RewardPurchase, Task, TaskInstance, UserSettings, WorkspaceSnapshot } from "../storage";

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
  return apiFetch<{ id: string; email: string; settings?: UserSettings }>("/me");
}

export async function listWorkspaces(): Promise<string[]> {
  const data = await apiFetch<{ workspaces: string[] }>("/workspaces");
  return data.workspaces ?? [];
}

export async function fetchWorkspaceBalance(workspaceId: string) {
  return apiFetch<{ workspace_id: string; balance: number }>(`/workspaces/${workspaceId}/balance`);
}

export async function createWorkspace(name: string, type: string) {
  return apiFetch<{ id: string }>("/workspaces", {
    method: "POST",
    body: JSON.stringify({ name, type })
  });
}

export async function createInvite(workspaceId: string) {
  return apiFetch<{ code: string }>(`/workspaces/${workspaceId}/invite`, { method: "POST" });
}

export async function acceptInvite(code: string) {
  return apiFetch<{ workspace_id: string }>("/invites/accept", {
    method: "POST",
    body: JSON.stringify({ code })
  });
}

export async function listWorkspaceMembers(workspaceId: string) {
  return apiFetch<{ members: { id: string; email: string; role: string; created_at: string }[] }>(
    `/workspaces/${workspaceId}/members`
  );
}

export async function addTask(
  workspaceId: string,
  form: {
    title: string;
    value: number;
    dueDate: string;
    is_recurring?: boolean;
    recurrence_weekdays?: number[];
    start_date?: string | null;
    end_date?: string | null;
    timezone?: string | null;
  }
): Promise<Task> {
  const payload = {
    workspace_id: workspaceId,
    title: form.title,
    value: form.value,
    due_date: form.dueDate || null,
    is_recurring: form.is_recurring ?? false,
    recurrence_weekdays: form.recurrence_weekdays ?? [],
    start_date: form.start_date ?? null,
    end_date: form.end_date ?? null,
    timezone: form.timezone ?? null,
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
    status: "open",
    is_recurring: form.is_recurring ?? false,
    recurrence_weekdays: form.recurrence_weekdays ?? []
  };
}

export async function completeTask(workspaceId: string, taskId: string, occurrenceDate?: string | null) {
  await apiFetch(`/tasks/${taskId}/complete`, {
    method: "POST",
    body: JSON.stringify({ workspace_id: workspaceId, occurrence_date: occurrenceDate ?? null })
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
  form: { title: string; cost: number; description: string; icon: string; one_time?: boolean }
): Promise<Reward> {
  const payload = {
    workspace_id: workspaceId,
    title: form.title,
    description: form.description,
    cost: form.cost,
    is_shared: false,
    one_time: form.one_time ?? false
  };
  const data = await apiFetch<{ id: string }>("/rewards", { method: "POST", body: JSON.stringify(payload) });
  return {
    id: data.id,
    workspace_id: workspaceId,
    title: form.title,
    description: form.description,
    cost: form.cost,
    icon: form.icon,
    one_time: form.one_time ?? false
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

export async function listTaskInstances(workspaceId: string, from: string, to: string) {
  return apiFetch<{ instances: TaskInstance[] }>(
    `/tasks?workspace_id=${workspaceId}&from=${from}&to=${to}`
  );
}

export async function listRewardPurchases(workspaceId: string) {
  return apiFetch<{ purchases: RewardPurchase[] }>(`/rewards/purchases?workspace_id=${workspaceId}`);
}

export async function updateSettings(settings: { theme: string; last_active_workspace?: string | null }) {
  await apiFetch("/settings", { method: "PUT", body: JSON.stringify(settings) });
}

function mergeEntities<T extends { id: string }>(current: T[], incoming: T[]): T[] {
  const map = new Map(current.map((item) => [item.id, item]));
  incoming.forEach((item) => map.set(item.id, item));
  return Array.from(map.values());
}
