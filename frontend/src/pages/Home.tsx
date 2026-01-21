import { useMemo, useState } from "react";
import { Task } from "../storage";
import { Card } from "../components/Card";
import { ProgressBar } from "../components/ProgressBar";

const focusCards = [
  { id: "day", label: "День", progress: 65 },
  { id: "week", label: "Неделя", progress: 42 },
  { id: "month", label: "Месяц", progress: 28 },
  { id: "year", label: "Год", progress: 12 }
];

export function Home({
  balance,
  tasks,
  onAdd,
  onComplete,
  onEdit,
  onSave,
  onDelete
}: {
  balance: number;
  tasks: Task[];
  onAdd: (form: { title: string; value: number; dueDate: string }) => void;
  onComplete: (task: Task) => void;
  onEdit: (task: Task, update: Partial<Task>) => void;
  onSave: (task: Task) => void;
  onDelete: (task: Task) => void;
}) {
  const todayTasks = useMemo(
    () => tasks.filter((task) => !task.deleted_at && task.status !== "done"),
    [tasks]
  );

  return (
    <div className="page">
      <header className="page__header">
        <div>
          <h1>Добро пожаловать</h1>
          <p className="muted">Сфокусируйтесь на важных целях</p>
        </div>
        <div className="balance-chip">
          <span>🔥</span>
          <strong>{balance.toFixed(0)}</strong>
        </div>
      </header>

      <div className="grid grid--2">
        {focusCards.map((card) => (
          <Card key={card.id} className="goal-card">
            <div className="goal-card__title">{card.label}</div>
            <ProgressBar value={card.progress} />
            <span className="muted">{card.progress}%</span>
          </Card>
        ))}
      </div>

      <Card className="streak-card">
        <div className="streak-card__header">
          <div>
            <h3>Полоса достижений</h3>
            <p className="muted">7 дней подряд</p>
          </div>
          <span className="streak-card__fire">🔥</span>
        </div>
        <ProgressBar value={72} />
      </Card>

      <Card>
        <div className="section-header">
          <h3>Задачи на сегодня</h3>
          <span className="muted">{todayTasks.length}</span>
        </div>
        <TaskForm onAdd={onAdd} />
        <div className="list">
          {todayTasks.map((task) => (
            <div key={task.id} className="list-item">
              <div className="list-item__icon">✅</div>
              <div className="list-item__body">
                <input
                  className="inline-input"
                  value={task.title}
                  onChange={(event) => onEdit(task, { title: event.target.value })}
                  onBlur={() => onSave(task)}
                />
                <span className="muted">+{task.value} 🔥</span>
              </div>
              <div className="list-item__trailing">
                <button className="ghost" onClick={() => onComplete(task)}>
                  Done
                </button>
                <button className="ghost" onClick={() => onDelete(task)}>
                  ✕
                </button>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}

function TaskForm({ onAdd }: { onAdd: (form: { title: string; value: number; dueDate: string }) => void }) {
  const [title, setTitle] = useState("");
  const [value, setValue] = useState(5);
  const [dueDate, setDueDate] = useState("");

  return (
    <div className="form-row">
      <input
        placeholder="Новая задача"
        value={title}
        onChange={(event) => setTitle(event.target.value)}
      />
      <input type="number" value={value} onChange={(event) => setValue(Number(event.target.value))} />
      <input type="date" value={dueDate} onChange={(event) => setDueDate(event.target.value)} />
      <button
        onClick={() => {
          if (!title) return;
          onAdd({ title, value, dueDate });
          setTitle("");
          setDueDate("");
        }}
      >
        Добавить
      </button>
    </div>
  );
}
