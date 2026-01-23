import { useMemo, useState } from "react";
import { Task, TaskInstance } from "../storage";
import { Card } from "../components/Card";
import { ProgressBar } from "../components/ProgressBar";

type PeriodStat = {
  id: string;
  label: string;
  done: number;
  total: number;
};

export function Home({
  balance,
  periodStats,
  todayTasks,
  streakDays,
  onAdd,
  onComplete,
  onEdit,
  onSave,
  onDelete
}: {
  balance: number;
  periodStats: PeriodStat[];
  todayTasks: TaskInstance[];
  streakDays: number;
  onAdd: (form: {
    title: string;
    value: number;
    dueDate: string;
    is_recurring?: boolean;
    recurrence_weekdays?: number[];
  }) => void;
  onComplete: (task: TaskInstance) => void;
  onEdit: (task: Task, update: Partial<Task>) => void;
  onSave: (task: Task) => void;
  onDelete: (task: Task) => void;
}) {
  const openTodayTasks = useMemo(() => todayTasks.filter((task) => !task.done), [todayTasks]);

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
        {periodStats.map((card) => {
          const percent = card.total === 0 ? 0 : Math.round((card.done / card.total) * 100);
          return (
            <Card key={card.id} className="goal-card">
              <div className="goal-card__title">{card.label}</div>
              <ProgressBar value={percent} />
              <span className="muted">
                {percent}% · {card.done}/{card.total}
              </span>
            </Card>
          );
        })}
      </div>

      <Card className="streak-card">
        <div className="streak-card__header">
          <div>
            <h3>Полоса достижений</h3>
            <p className="muted">{streakDays} дней подряд</p>
          </div>
          <span className="streak-card__fire">🔥</span>
        </div>
        <ProgressBar value={Math.min(100, streakDays * 10)} />
      </Card>

      <Card>
        <div className="section-header">
          <h3>Задачи на сегодня</h3>
          <span className="muted">{openTodayTasks.length}</span>
        </div>
        <TaskForm onAdd={onAdd} />
        <div className="list">
          {todayTasks.map((task) => (
            <div key={`${task.id}-${task.occurrence_date}`} className="list-item">
              <div className="list-item__icon">{task.done ? "✅" : "⏳"}</div>
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
                <button className="ghost" onClick={() => onComplete(task)} disabled={task.done}>
                  {task.done ? "Готово" : "Done"}
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

function TaskForm({
  onAdd
}: {
  onAdd: (form: {
    title: string;
    value: number;
    dueDate: string;
    is_recurring?: boolean;
    recurrence_weekdays?: number[];
  }) => void;
}) {
  const [title, setTitle] = useState("");
  const [value, setValue] = useState(5);
  const [dueDate, setDueDate] = useState("");
  const [isRecurring, setIsRecurring] = useState(false);
  const [weekdays, setWeekdays] = useState<number[]>([]);

  const toggleWeekday = (day: number) => {
    setWeekdays((prev) => (prev.includes(day) ? prev.filter((item) => item !== day) : [...prev, day]));
  };

  return (
    <div className="form-row">
      <input
        placeholder="Новая задача"
        value={title}
        onChange={(event) => setTitle(event.target.value)}
      />
      <input type="number" value={value} onChange={(event) => setValue(Number(event.target.value))} />
      <input type="date" value={dueDate} onChange={(event) => setDueDate(event.target.value)} />
      <label className="checkbox">
        <input type="checkbox" checked={isRecurring} onChange={(event) => setIsRecurring(event.target.checked)} />
        Повторять
      </label>
      {isRecurring && (
        <div className="weekday-picker">
          {["Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"].map((label, idx) => (
            <button
              key={label}
              type="button"
              className={weekdays.includes(idx) ? "weekday active" : "weekday"}
              onClick={() => toggleWeekday(idx)}
            >
              {label}
            </button>
          ))}
        </div>
      )}
      <button
        onClick={() => {
          if (!title) return;
          onAdd({ title, value, dueDate, is_recurring: isRecurring, recurrence_weekdays: weekdays });
          setTitle("");
          setDueDate("");
          setIsRecurring(false);
          setWeekdays([]);
        }}
      >
        Добавить
      </button>
    </div>
  );
}
