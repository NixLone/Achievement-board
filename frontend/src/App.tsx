import { useEffect, useMemo, useState } from "react";
import {
  addAchievement,
  addReward,
  addTask,
  buyReward,
  completeTask,
  fetchWorkspaceBalance,
  getMe,
  listWorkspaces,
  login,
  register,
  runSyncPull,
  updateLocalCache
} from "./api";
import {
  Achievement,
  Reward,
  Task,
  WorkspaceSnapshot,
  emptySnapshot,
  loadSnapshot,
  saveSnapshot
} from "./storage";

const tabs = ["Today", "Plans", "Store", "Achievements", "Settings"] as const;

type Tab = (typeof tabs)[number];

export default function App() {
  const [activeTab, setActiveTab] = useState<Tab>("Today");
  const [snapshot, setSnapshot] = useState<WorkspaceSnapshot>(emptySnapshot());
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [status, setStatus] = useState<string | null>(null);
  const [balance, setBalance] = useState<number>(0);

  useEffect(() => {
    loadSnapshot().then((data) => {
      setSnapshot(data);
    });
  }, []);

  const todayTasks = useMemo(
    () => snapshot.tasks.filter((task) => !task.deleted_at && task.status !== "done"),
    [snapshot.tasks]
  );

  async function handleAuth(action: "login" | "register") {
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
      await saveSnapshot(nextSnapshot);
      await refreshBalance(workspaceID);
      setStatus("Авторизация успешна");
    } catch (error) {
      setStatus("Не удалось авторизоваться");
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
      await saveSnapshot(updated);
      await refreshBalance(updated.workspaceId);
      setStatus("Синхронизация завершена");
    } catch (error) {
      setStatus("Ошибка синхронизации");
    }
  }

  async function refreshBalance(workspaceId: string) {
    const data = await fetchWorkspaceBalance(workspaceId);
    setBalance(data.balance);
  }

  async function handleAddTask(form: { title: string; value: number; dueDate: string }) {
    if (!snapshot.workspaceId) return;
    const task = await addTask(snapshot.workspaceId, form);
    const updated = updateLocalCache(snapshot, { tasks: [task] });
    setSnapshot(updated);
    await saveSnapshot(updated);
  }

  async function handleCompleteTask(task: Task) {
    if (!snapshot.workspaceId) return;
    await completeTask(snapshot.workspaceId, task.id);
    const updated = updateLocalCache(snapshot, {
      tasks: snapshot.tasks.map((item) =>
        item.id === task.id
          ? { ...item, status: "done", done_at: new Date().toISOString() }
          : item
      )
    });
    setSnapshot(updated);
    await saveSnapshot(updated);
    await refreshBalance(snapshot.workspaceId);
  }

  async function handleAddReward(form: { title: string; cost: number; description: string }) {
    if (!snapshot.workspaceId) return;
    const reward = await addReward(snapshot.workspaceId, form);
    const updated = updateLocalCache(snapshot, { rewards: [reward] });
    setSnapshot(updated);
    await saveSnapshot(updated);
  }

  async function handleBuyReward(reward: Reward) {
    if (!snapshot.workspaceId) return;
    await buyReward(snapshot.workspaceId, reward.id);
    await refreshBalance(snapshot.workspaceId);
  }

  async function handleAddAchievement(form: { title: string; description: string }) {
    if (!snapshot.workspaceId) return;
    const achievement = await addAchievement(snapshot.workspaceId, form);
    const updated = updateLocalCache(snapshot, { achievements: [achievement] });
    setSnapshot(updated);
    await saveSnapshot(updated);
  }

  return (
    <div className="app">
      <header className="app__header">
        <div>
          <h1>FireGoals</h1>
          <p className="muted">Цели, задачи и огоньки</p>
        </div>
        <div className="balance">
          <span>Огоньки</span>
          <strong>{balance.toFixed(2)}</strong>
        </div>
      </header>
      <nav className="app__nav">
        {tabs.map((tab) => (
          <button
            key={tab}
            className={tab === activeTab ? "active" : ""}
            onClick={() => setActiveTab(tab)}
          >
            {tab}
          </button>
        ))}
      </nav>
      <main className="app__main">
        {activeTab === "Today" && (
          <section className="card">
            <h2>Сегодня</h2>
            <TaskForm onAdd={handleAddTask} />
            <ul className="list">
              {todayTasks.map((task) => (
                <li key={task.id} className="list__item">
                  <div>
                    <p>{task.title}</p>
                    <span className="muted">+{task.value} огоньков</span>
                  </div>
                  <button onClick={() => handleCompleteTask(task)}>Готово</button>
                </li>
              ))}
            </ul>
          </section>
        )}
        {activeTab === "Plans" && (
          <section className="card">
            <h2>Планы</h2>
            <p className="muted">Добавляйте цели и задачи на неделю, месяц или год.</p>
            <p>В скором времени появится детальная планировка.</p>
          </section>
        )}
        {activeTab === "Store" && (
          <section className="card">
            <h2>Магазин наград</h2>
            <RewardForm onAdd={handleAddReward} />
            <ul className="list">
              {snapshot.rewards.filter((reward) => !reward.deleted_at).map((reward) => (
                <li key={reward.id} className="list__item">
                  <div>
                    <p>{reward.title}</p>
                    <span className="muted">{reward.cost} огоньков</span>
                  </div>
                  <button onClick={() => handleBuyReward(reward)}>Купить</button>
                </li>
              ))}
            </ul>
          </section>
        )}
        {activeTab === "Achievements" && (
          <section className="card">
            <h2>Достижения</h2>
            <AchievementForm onAdd={handleAddAchievement} />
            <ul className="list">
              {snapshot.achievements
                .filter((achievement) => !achievement.deleted_at)
                .map((achievement) => (
                <li key={achievement.id} className="list__item">
                  <div>
                    <p>{achievement.title}</p>
                    <span className="muted">{achievement.description}</span>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        )}
        {activeTab === "Settings" && (
          <section className="card">
            <h2>Профиль и синхронизация</h2>
            <div className="auth">
              <input
                type="email"
                placeholder="Email"
                value={email}
                onChange={(event) => setEmail(event.target.value)}
              />
              <input
                type="password"
                placeholder="Пароль"
                value={password}
                onChange={(event) => setPassword(event.target.value)}
              />
              <div className="auth__actions">
                <button onClick={() => handleAuth("login")}>Войти</button>
                <button className="secondary" onClick={() => handleAuth("register")}>
                  Регистрация
                </button>
              </div>
            </div>
            <button className="full" onClick={handleSync}>
              Синхронизировать сейчас
            </button>
            {status && <p className="muted">{status}</p>}
          </section>
        )}
      </main>
    </div>
  );
}

function TaskForm({ onAdd }: { onAdd: (data: { title: string; value: number; dueDate: string }) => void }) {
  const [title, setTitle] = useState("");
  const [value, setValue] = useState(10);
  const [dueDate, setDueDate] = useState("");

  return (
    <div className="form">
      <input
        value={title}
        placeholder="Название задачи"
        onChange={(event) => setTitle(event.target.value)}
      />
      <input
        type="number"
        value={value}
        onChange={(event) => setValue(Number(event.target.value))}
      />
      <input type="date" value={dueDate} onChange={(event) => setDueDate(event.target.value)} />
      <button
        onClick={() => {
          if (title) {
            onAdd({ title, value, dueDate });
            setTitle("");
            setDueDate("");
          }
        }}
      >
        Добавить
      </button>
    </div>
  );
}

function RewardForm({
  onAdd
}: {
  onAdd: (data: { title: string; cost: number; description: string }) => void;
}) {
  const [title, setTitle] = useState("");
  const [cost, setCost] = useState(30);
  const [description, setDescription] = useState("");

  return (
    <div className="form">
      <input
        value={title}
        placeholder="Награда"
        onChange={(event) => setTitle(event.target.value)}
      />
      <input type="number" value={cost} onChange={(event) => setCost(Number(event.target.value))} />
      <input
        value={description}
        placeholder="Описание"
        onChange={(event) => setDescription(event.target.value)}
      />
      <button
        onClick={() => {
          if (title) {
            onAdd({ title, cost, description });
            setTitle("");
            setDescription("");
          }
        }}
      >
        Добавить
      </button>
    </div>
  );
}

function AchievementForm({
  onAdd
}: {
  onAdd: (data: { title: string; description: string }) => void;
}) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");

  return (
    <div className="form">
      <input
        value={title}
        placeholder="Достижение"
        onChange={(event) => setTitle(event.target.value)}
      />
      <input
        value={description}
        placeholder="Описание"
        onChange={(event) => setDescription(event.target.value)}
      />
      <button
        onClick={() => {
          if (title) {
            onAdd({ title, description });
            setTitle("");
            setDescription("");
          }
        }}
      >
        Добавить
      </button>
    </div>
  );
}
