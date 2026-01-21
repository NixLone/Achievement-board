import { useMemo, useState } from "react";
import { BottomNav, TabId } from "./components/BottomNav";
import { Home } from "./pages/Home";
import { Plans } from "./pages/Plans";
import { Store } from "./pages/Store";
import { Achievements } from "./pages/Achievements";
import { Settings } from "./pages/Settings";
import {
  addAchievement,
  addReward,
  addTask,
  buyReward,
  completeTask,
  deleteAchievement,
  deleteReward,
  deleteTask,
  fetchWorkspaceBalance,
  getMe,
  listWorkspaces,
  login,
  register,
  runSyncPull,
  updateAchievement,
  updateReward,
  updateTask
} from "./api";
import { ApiError, hasApiBaseUrl } from "./api/client";
import { useStore } from "./state/store";
import { Achievement, Reward, Task } from "./storage";
import { mergeById } from "./utils/merge";
import { useTheme } from "./theme/useTheme";

export default function App() {
  const { snapshot, setSnapshot } = useStore();
  const [activeTab, setActiveTab] = useState<TabId>("home");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [status, setStatus] = useState<string | null>(null);
  const [balance, setBalance] = useState<number>(0);
  const { theme, setTheme } = useTheme();

  const apiMissing = !hasApiBaseUrl();

  const tasks = useMemo(() => snapshot.tasks ?? [], [snapshot.tasks]);
  const rewards = useMemo(() => snapshot.rewards ?? [], [snapshot.rewards]);
  const achievements = useMemo(() => snapshot.achievements ?? [], [snapshot.achievements]);

  async function refreshBalance(workspaceId?: string | null) {
    if (!workspaceId) return;
    const data = await fetchWorkspaceBalance(workspaceId);
    setBalance(data.balance);
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
      const workspaceID = await listWorkspaces();
      const nextSnapshot = { ...snapshot, user: me, workspaceId: workspaceID };
      setSnapshot(nextSnapshot);
      await refreshBalance(workspaceID);
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
      await refreshBalance(updated.workspaceId);
      setStatus("Синхронизация завершена");
    } catch (error) {
      if (error instanceof ApiError) {
        setStatus(error.message);
      } else {
        setStatus("Ошибка синхронизации");
      }
    }
  }

  async function handleAddTask(form: { title: string; value: number; dueDate: string }) {
    if (!snapshot.workspaceId) return;
    try {
      const task = await addTask(snapshot.workspaceId, form);
      const updated = { ...snapshot, tasks: mergeById(snapshot.tasks, [task]) };
      setSnapshot(updated);
    } catch (error) {
      setStatus("Не удалось добавить задачу");
    }
  }

  async function handleCompleteTask(task: Task) {
    if (!snapshot.workspaceId) return;
    await completeTask(snapshot.workspaceId, task.id);
    const updatedTask = { ...task, status: "done", done_at: new Date().toISOString() };
    const updated = { ...snapshot, tasks: mergeById(snapshot.tasks, [updatedTask]) };
    setSnapshot(updated);
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
    } catch (error) {
      setStatus("Не удалось удалить задачу");
    }
  }

  async function handleAddReward(form: { title: string; cost: number; description: string; icon: string }) {
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
        cost: reward.cost
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

  return (
    <div className="app">
      {status && <div className="toast">{status}</div>}
      {activeTab === "home" && (
        <Home
          balance={balance}
          tasks={tasks}
          onAdd={handleAddTask}
          onComplete={handleCompleteTask}
          onEdit={handleEditTask}
          onSave={handleSaveTask}
          onDelete={handleDeleteTask}
        />
      )}
      {activeTab === "plans" && <Plans />}
      {activeTab === "store" && (
        <Store
          rewards={rewards}
          onBuy={handleBuyReward}
          onAdd={handleAddReward}
          onEdit={handleEditReward}
          onSave={handleSaveReward}
          onDelete={handleDeleteReward}
        />
      )}
      {activeTab === "achievements" && (
        <Achievements
          achievements={achievements}
          onAdd={handleAddAchievement}
          onEdit={handleEditAchievement}
          onSave={handleSaveAchievement}
          onDelete={handleDeleteAchievement}
        />
      )}
      {activeTab === "settings" && (
        <Settings
          email={email}
          password={password}
          userEmail={snapshot.user?.email ?? null}
          workspaceId={snapshot.workspaceId ?? null}
          theme={theme}
          onThemeChange={setTheme}
          onEmailChange={setEmail}
          onPasswordChange={setPassword}
          onLogin={() => handleAuth("login")}
          onRegister={() => handleAuth("register")}
          onSync={handleSync}
          status={status}
          apiMissing={apiMissing}
          lastSync={snapshot.lastSync ?? null}
        />
      )}
      <BottomNav active={activeTab} onChange={setActiveTab} />
    </div>
  );
}
