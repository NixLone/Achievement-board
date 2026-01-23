import { useEffect, useMemo, useState } from "react";
import { BottomNav, TabId } from "./components/BottomNav";
import { Home } from "./pages/Home";
import { Plans } from "./pages/Plans";
import { Store } from "./pages/Store";
import { Achievements } from "./pages/Achievements";
import { Settings } from "./pages/Settings";
import { Workspace } from "./pages/Workspace";
import {
  acceptInvite,
  addAchievement,
  addReward,
  addTask,
  buyReward,
  completeTask,
  createInvite,
  createWorkspace,
  deleteAchievement,
  deleteReward,
  deleteTask,
  fetchWorkspaceBalance,
  getMe,
  listRewardPurchases,
  listTaskInstances,
  listWorkspaceMembers,
  listWorkspaces,
  login,
  register,
  runSyncPull,
  updateAchievement,
  updateReward,
  updateSettings,
  updateTask
} from "./api";
import { ApiError, hasApiBaseUrl } from "./api/client";
import { useStore } from "./state/store";
import { Achievement, Reward, RewardPurchase, Task, TaskInstance } from "./storage";
import { mergeById } from "./utils/merge";
import { useTheme } from "./theme/useTheme";

function formatDate(date: Date) {
  return date.toISOString().slice(0, 10);
}

function getWeekRange(base: Date) {
  const day = base.getDay();
  const diff = (day === 0 ? -6 : 1) - day;
  const monday = new Date(base);
  monday.setDate(base.getDate() + diff);
  const sunday = new Date(monday);
  sunday.setDate(monday.getDate() + 6);
  return { start: monday, end: sunday };
}

export default function App() {
  const { snapshot, setSnapshot } = useStore();
  const [activeTab, setActiveTab] = useState<TabId>("home");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [status, setStatus] = useState<string | null>(null);
  const [balance, setBalance] = useState<number>(0);
  const [workspaces, setWorkspaces] = useState<string[]>([]);
  const [purchases, setPurchases] = useState<RewardPurchase[]>([]);
  const [members, setMembers] = useState<{ id: string; email: string; role: string }[]>([]);
  const [dayInstances, setDayInstances] = useState<TaskInstance[]>([]);
  const [weekInstances, setWeekInstances] = useState<TaskInstance[]>([]);
  const [monthInstances, setMonthInstances] = useState<TaskInstance[]>([]);
  const [yearInstances, setYearInstances] = useState<TaskInstance[]>([]);
  const { theme, setTheme } = useTheme();

  const apiMissing = !hasApiBaseUrl();

  const tasks = useMemo(() => snapshot.tasks ?? [], [snapshot.tasks]);
  const rewards = useMemo(() => snapshot.rewards ?? [], [snapshot.rewards]);
  const achievements = useMemo(() => snapshot.achievements ?? [], [snapshot.achievements]);

  const periodStats = useMemo(() => {
    const day = summarizeInstances(dayInstances);
    const week = summarizeInstances(weekInstances);
    const month = summarizeInstances(monthInstances);
    const year = summarizeInstances(yearInstances);
    return [
      { id: "day", label: "День", ...day },
      { id: "week", label: "Неделя", ...week },
      { id: "month", label: "Месяц", ...month },
      { id: "year", label: "Год", ...year }
    ];
  }, [dayInstances, weekInstances, monthInstances, yearInstances]);

  const streakDays = useMemo(() => {
    const today = new Date();
    let streak = 0;
    for (let i = 0; i < 7; i += 1) {
      const date = new Date(today);
      date.setDate(today.getDate() - i);
      const dateKey = formatDate(date);
      const tasksForDay = weekInstances.filter((task) => task.occurrence_date === dateKey);
      if (tasksForDay.length === 0) break;
      if (!tasksForDay.every((task) => task.done)) break;
      streak += 1;
    }
    return streak;
  }, [weekInstances]);

  useEffect(() => {
    if (!snapshot.workspaceId) return;
    refreshWorkspace(snapshot.workspaceId).catch(() => undefined);
  }, [snapshot.workspaceId]);

  async function refreshWorkspace(workspaceId: string) {
    await refreshBalance(workspaceId);
    await refreshInstances(workspaceId);
    await refreshPurchases(workspaceId);
    await refreshMembers(workspaceId);
  }

  async function refreshBalance(workspaceId?: string | null) {
    if (!workspaceId) return;
    const data = await fetchWorkspaceBalance(workspaceId);
    setBalance(data.balance);
  }

  async function refreshInstances(workspaceId: string) {
    const today = new Date();
    const dayKey = formatDate(today);
    const weekRange = getWeekRange(today);
    const monthStart = new Date(today.getFullYear(), today.getMonth(), 1);
    const monthEnd = new Date(today.getFullYear(), today.getMonth() + 1, 0);
    const yearStart = new Date(today.getFullYear(), 0, 1);
    const yearEnd = new Date(today.getFullYear(), 11, 31);

    const dayData = await listTaskInstances(workspaceId, dayKey, dayKey);
    const weekData = await listTaskInstances(workspaceId, formatDate(weekRange.start), formatDate(weekRange.end));
    const monthData = await listTaskInstances(workspaceId, formatDate(monthStart), formatDate(monthEnd));
    const yearData = await listTaskInstances(workspaceId, formatDate(yearStart), formatDate(yearEnd));

    setDayInstances(dayData.instances ?? []);
    setWeekInstances(weekData.instances ?? []);
    setMonthInstances(monthData.instances ?? []);
    setYearInstances(yearData.instances ?? []);
  }

  async function refreshPurchases(workspaceId: string) {
    const data = await listRewardPurchases(workspaceId);
    setPurchases(data.purchases ?? []);
  }

  async function refreshMembers(workspaceId: string) {
    const data = await listWorkspaceMembers(workspaceId);
    setMembers(data.members ?? []);
  }

  async function handleAuth(action: "login" | "register") {
    if (apiMissing) {
      setStatus("API URL не настроен");
      return;
    }
    setStatus(null);
    try {
      if (action === "register") {
        await register(email, password);
      }
      await login(email, password);
      const me = await getMe();
      const workspaceList = await listWorkspaces();
      const activeWorkspace = me.settings?.last_active_workspace ?? workspaceList[0] ?? null;
      const nextSnapshot = { ...snapshot, user: { id: me.id, email: me.email }, workspaceId: activeWorkspace, settings: me.settings };
      setSnapshot(nextSnapshot);
      setWorkspaces(workspaceList);
      if (me.settings?.theme) {
        setTheme(me.settings.theme);
      }
      if (activeWorkspace) {
        await refreshWorkspace(activeWorkspace);
      }
      setStatus("Авторизация успешна");
    } catch (error) {
      if (error instanceof ApiError) {
        setStatus(error.message);
      } else {
        setStatus("Сервер недоступен");
      }
    }
  }

  async function handleSync() {
    if (!snapshot.workspaceId) {
      setStatus("Сначала войдите в аккаунт");
      return;
    }
    setStatus("Синхронизация...");
    try {
      const updated = await runSyncPull(snapshot);
      setSnapshot(updated);
      if (updated.workspaceId) {
        await refreshWorkspace(updated.workspaceId);
      }
      setStatus("Синхронизация завершена");
    } catch (error) {
      if (error instanceof ApiError) {
        setStatus(error.message);
      } else {
        setStatus("Ошибка синхронизации");
      }
    }
  }

  async function handleAddTask(form: {
    title: string;
    value: number;
    dueDate: string;
    is_recurring?: boolean;
    recurrence_weekdays?: number[];
  }) {
    if (!snapshot.workspaceId) return;
    try {
      const task = await addTask(snapshot.workspaceId, form);
      const updated = { ...snapshot, tasks: mergeById(snapshot.tasks, [task]) };
      setSnapshot(updated);
      await refreshInstances(snapshot.workspaceId);
    } catch (error) {
      setStatus("Не удалось добавить задачу");
    }
  }

  async function handleCompleteTask(task: TaskInstance) {
    if (!snapshot.workspaceId) return;
    await completeTask(snapshot.workspaceId, task.id, task.occurrence_date);
    await refreshInstances(snapshot.workspaceId);
    await refreshBalance(snapshot.workspaceId);
  }

  async function handleEditTask(task: Task, update: Partial<Task>) {
    if (!snapshot.workspaceId) return;
    const updatedTask = { ...task, ...update };
    setSnapshot({ ...snapshot, tasks: mergeById(snapshot.tasks, [updatedTask]) });
  }

  async function handleSaveTask(task: Task) {
    if (!snapshot.workspaceId) return;
    try {
      await updateTask(snapshot.workspaceId, task.id, {
        title: task.title,
        description: task.description,
        value: task.value
      });
    } catch (error) {
      setStatus("Не удалось сохранить задачу");
    }
  }

  async function handleDeleteTask(task: Task) {
    if (!snapshot.workspaceId) return;
    const updatedTask = { ...task, deleted_at: new Date().toISOString() };
    setSnapshot({ ...snapshot, tasks: mergeById(snapshot.tasks, [updatedTask]) });
    try {
      await deleteTask(snapshot.workspaceId, task.id);
      await refreshInstances(snapshot.workspaceId);
    } catch (error) {
      setStatus("Не удалось удалить задачу");
    }
  }

  async function handleAddReward(form: { title: string; cost: number; description: string; icon: string; one_time: boolean }) {
    if (!snapshot.workspaceId) return;
    try {
      const reward = await addReward(snapshot.workspaceId, form);
      const updated = { ...snapshot, rewards: mergeById(snapshot.rewards, [reward]) };
      setSnapshot(updated);
    } catch (error) {
      setStatus("Не удалось добавить награду");
    }
  }

  async function handleBuyReward(reward: Reward) {
    if (!snapshot.workspaceId) return;
    try {
      await buyReward(snapshot.workspaceId, reward.id);
      await refreshBalance(snapshot.workspaceId);
      await refreshPurchases(snapshot.workspaceId);
    } catch (error) {
      if (error instanceof ApiError) {
        setStatus(error.message);
      }
    }
  }

  async function handleEditReward(reward: Reward, update: Partial<Reward>) {
    if (!snapshot.workspaceId) return;
    const updatedReward = { ...reward, ...update };
    setSnapshot({ ...snapshot, rewards: mergeById(snapshot.rewards, [updatedReward]) });
  }

  async function handleSaveReward(reward: Reward) {
    if (!snapshot.workspaceId) return;
    try {
      await updateReward(snapshot.workspaceId, reward.id, {
        title: reward.title,
        description: reward.description,
        cost: reward.cost,
        one_time: reward.one_time
      });
    } catch (error) {
      setStatus("Не удалось сохранить награду");
    }
  }

  async function handleDeleteReward(reward: Reward) {
    if (!snapshot.workspaceId) return;
    const updatedReward = { ...reward, deleted_at: new Date().toISOString() };
    setSnapshot({ ...snapshot, rewards: mergeById(snapshot.rewards, [updatedReward]) });
    try {
      await deleteReward(snapshot.workspaceId, reward.id);
    } catch (error) {
      setStatus("Не удалось удалить награду");
    }
  }

  async function handleAddAchievement(form: { title: string; description: string; imageUrl: string }) {
    if (!snapshot.workspaceId) return;
    try {
      const achievement = await addAchievement(snapshot.workspaceId, form);
      const updated = { ...snapshot, achievements: mergeById(snapshot.achievements, [achievement]) };
      setSnapshot(updated);
    } catch (error) {
      setStatus("Не удалось добавить достижение");
    }
  }

  async function handleEditAchievement(achievement: Achievement, update: Partial<Achievement>) {
    if (!snapshot.workspaceId) return;
    const updatedAchievement = { ...achievement, ...update };
    setSnapshot({ ...snapshot, achievements: mergeById(snapshot.achievements, [updatedAchievement]) });
  }

  async function handleSaveAchievement(achievement: Achievement) {
    if (!snapshot.workspaceId) return;
    try {
      await updateAchievement(snapshot.workspaceId, achievement.id, {
        title: achievement.title,
        description: achievement.description,
        image_url: achievement.image_url
      });
    } catch (error) {
      setStatus("Не удалось сохранить достижение");
    }
  }

  async function handleDeleteAchievement(achievement: Achievement) {
    if (!snapshot.workspaceId) return;
    const updatedAchievement = { ...achievement, deleted_at: new Date().toISOString() };
    setSnapshot({ ...snapshot, achievements: mergeById(snapshot.achievements, [updatedAchievement]) });
    try {
      await deleteAchievement(snapshot.workspaceId, achievement.id);
    } catch (error) {
      setStatus("Не удалось удалить достижение");
    }
  }

  async function handleThemeChange(nextTheme: string) {
    setTheme(nextTheme as typeof theme);
    if (!snapshot.user?.id || apiMissing) return;
    try {
      await updateSettings({ theme: nextTheme, last_active_workspace: snapshot.workspaceId ?? null });
    } catch (error) {
      setStatus("Не удалось сохранить тему");
    }
  }

  async function handleWorkspaceChange(nextWorkspaceId: string) {
    if (!nextWorkspaceId) return;
    setSnapshot({ ...snapshot, workspaceId: nextWorkspaceId });
    if (!apiMissing) {
      await updateSettings({ theme, last_active_workspace: nextWorkspaceId });
    }
    await refreshWorkspace(nextWorkspaceId);
  }

  async function handleCreateWorkspace(name: string) {
    const result = await createWorkspace(name, "shared");
    const workspaceList = await listWorkspaces();
    setWorkspaces(workspaceList);
    if (result.id) {
      await handleWorkspaceChange(result.id);
    }
  }

  async function handleCreateInvite() {
    if (!snapshot.workspaceId) return;
    try {
      const result = await createInvite(snapshot.workspaceId);
      setStatus(`Инвайт создан: ${result.code}`);
    } catch (error) {
      setStatus("Не удалось создать инвайт");
    }
  }

  async function handleAcceptInvite(code: string) {
    try {
      const result = await acceptInvite(code);
      const workspaceList = await listWorkspaces();
      setWorkspaces(workspaceList);
      if (result.workspace_id) {
        await handleWorkspaceChange(result.workspace_id);
      }
    } catch (error) {
      setStatus("Не удалось принять инвайт");
    }
  }

  async function handleRefreshMembers() {
    if (!snapshot.workspaceId) return;
    await refreshMembers(snapshot.workspaceId);
  }

  const content = (() => {
    switch (activeTab) {
      case "home":
        return (
          <Home
            balance={balance}
            periodStats={periodStats}
            todayTasks={dayInstances}
            streakDays={streakDays}
            onAdd={handleAddTask}
            onComplete={handleCompleteTask}
            onEdit={handleEditTask}
            onSave={handleSaveTask}
            onDelete={handleDeleteTask}
          />
        );
      case "plans":
        return (
          <Plans
            periodStats={periodStats}
            weekInstances={weekInstances}
            onAdd={handleAddTask}
            onComplete={handleCompleteTask}
          />
        );
      case "store":
        return (
          <Store
            rewards={rewards}
            purchases={purchases}
            onBuy={handleBuyReward}
            onAdd={handleAddReward}
            onEdit={handleEditReward}
            onSave={handleSaveReward}
            onDelete={handleDeleteReward}
          />
        );
      case "achievements":
        return (
          <Achievements
            achievements={achievements}
            onAdd={handleAddAchievement}
            onEdit={handleEditAchievement}
            onSave={handleSaveAchievement}
            onDelete={handleDeleteAchievement}
          />
        );
      case "workspace":
        return (
          <Workspace
            workspaceId={snapshot.workspaceId}
            members={members}
            onCreateWorkspace={handleCreateWorkspace}
            onCreateInvite={handleCreateInvite}
            onAcceptInvite={handleAcceptInvite}
            onRefresh={handleRefreshMembers}
          />
        );
      case "settings":
        return (
          <Settings
            email={email}
            password={password}
            userEmail={snapshot.user?.email}
            workspaceId={snapshot.workspaceId}
            workspaces={workspaces}
            theme={theme}
            onEmailChange={setEmail}
            onPasswordChange={setPassword}
            onThemeChange={handleThemeChange}
            onWorkspaceChange={handleWorkspaceChange}
            onLogin={() => handleAuth("login")}
            onRegister={() => handleAuth("register")}
            onSync={handleSync}
            status={status}
            apiMissing={apiMissing}
            lastSync={snapshot.lastSync}
          />
        );
      default:
        return null;
    }
  })();

  return (
    <div className="app">
      {content}
      <BottomNav active={activeTab} onChange={setActiveTab} />
    </div>
  );
}

function summarizeInstances(instances: TaskInstance[]) {
  const total = instances.length;
  const done = instances.filter((task) => task.done).length;
  return { done, total };
}
