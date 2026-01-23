import { useMemo, useState } from "react";
import { Card } from "../components/Card";
import { ProgressBar } from "../components/ProgressBar";
import { TaskInstance } from "../storage";

type PeriodStat = {
  id: string;
  label: string;
  done: number;
  total: number;
};

type DaySummary = {
  date: string;
  label: string;
  done: number;
  total: number;
};

const periods = ["День", "Неделя", "Месяц", "Год"] as const;
const weekdayLabels = ["Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"];

export function Plans({
  periodStats,
  weekInstances,
  onAdd,
  onComplete
}: {
  periodStats: PeriodStat[];
  weekInstances: TaskInstance[];
  onAdd: (form: { title: string; value: number; dueDate: string }) => void;
  onComplete: (task: TaskInstance) => void;
}) {
  const [active, setActive] = useState<(typeof periods)[number]>("Неделя");
  const summaries = useMemo(() => buildWeekSummaries(weekInstances, new Date()), [weekInstances]);
  const [selectedDay, setSelectedDay] = useState<DaySummary | null>(summaries[0] ?? null);

  const dayTasks = useMemo(() => {
    if (!selectedDay) return [];
    return weekInstances.filter((task) => task.occurrence_date === selectedDay.date);
  }, [weekInstances, selectedDay]);

  const activeStats = periodStats.find((stat) => {
    if (active === "День") return stat.id === "day";
    if (active === "Неделя") return stat.id === "week";
    if (active === "Месяц") return stat.id === "month";
    return stat.id === "year";
  });

  return (
    <div className="page">
      <header className="page__header">
        <h1>Планы</h1>
        <p className="muted">Фокус на горизонтах</p>
      </header>

      <div className="tabs">
        {periods.map((period) => (
          <button
            key={period}
            className={active === period ? "active" : ""}
            onClick={() => setActive(period)}
          >
            {period}
          </button>
        ))}
      </div>

      {activeStats && (
        <Card>
          <div className="section-header">
            <h3>{active} прогресс</h3>
            <span className="muted">
              {activeStats.done}/{activeStats.total}
            </span>
          </div>
          <ProgressBar value={activeStats.total === 0 ? 0 : Math.round((activeStats.done / activeStats.total) * 100)} />
        </Card>
      )}

      {active === "Неделя" && (
        <>
          <div className="grid grid--2">
            {summaries.map((summary) => {
              const percent = summary.total === 0 ? 0 : Math.round((summary.done / summary.total) * 100);
              return (
                <Card key={summary.date} className="weekday-card" onClick={() => setSelectedDay(summary)}>
                  <div className="weekday-card__header">
                    <span>{summary.label}</span>
                    <span className="muted">{summary.date}</span>
                  </div>
                  <ProgressBar value={percent} />
                  <div className="muted">
                    {percent}% · {summary.done}/{summary.total}
                  </div>
                </Card>
              );
            })}
          </div>

          <Card>
            <div className="section-header">
              <h3>Задачи на {selectedDay?.label}</h3>
              <span className="muted">{dayTasks.filter((task) => !task.done).length}</span>
            </div>
            <TaskForm
              onAdd={(form) => {
                if (!selectedDay) return;
                onAdd({ ...form, dueDate: selectedDay.date });
              }}
            />
            <div className="list">
              {dayTasks.map((task) => (
                <div key={`${task.id}-${task.occurrence_date}`} className="list-item">
                  <div className="list-item__icon">{task.done ? "✅" : "📌"}</div>
                  <div className="list-item__body">
                    <div>{task.title}</div>
                    <span className="muted">+{task.value} 🔥</span>
                  </div>
                  <div className="list-item__trailing">
                    <button className="ghost" onClick={() => onComplete(task)} disabled={task.done}>
                      {task.done ? "Готово" : "Done"}
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </>
      )}
    </div>
  );
}

function buildWeekSummaries(instances: TaskInstance[], baseDate: Date): DaySummary[] {
  const map = new Map<string, { done: number; total: number }>();
  instances.forEach((task) => {
    const entry = map.get(task.occurrence_date) ?? { done: 0, total: 0 };
    entry.total += 1;
    if (task.done) entry.done += 1;
    map.set(task.occurrence_date, entry);
  });
  const start = new Date(baseDate);
  const day = start.getDay();
  const diff = (day === 0 ? -6 : 1) - day;
  start.setDate(baseDate.getDate() + diff);

  return Array.from({ length: 7 }).map((_, idx) => {
    const date = new Date(start);
    date.setDate(start.getDate() + idx);
    const dateKey = date.toISOString().slice(0, 10);
    const value = map.get(dateKey) ?? { done: 0, total: 0 };
    return { date: dateKey, label: weekdayLabels[idx], done: value.done, total: value.total };
  });
}

function TaskForm({ onAdd }: { onAdd: (form: { title: string; value: number; dueDate: string }) => void }) {
  const [title, setTitle] = useState("");
  const [value, setValue] = useState(5);

  return (
    <div className="form-row">
      <input placeholder="Задача" value={title} onChange={(event) => setTitle(event.target.value)} />
      <input type="number" value={value} onChange={(event) => setValue(Number(event.target.value))} />
      <button
        onClick={() => {
          if (!title) return;
          onAdd({ title, value, dueDate: "" });
          setTitle("");
        }}
      >
        Добавить
      </button>
    </div>
  );
}
