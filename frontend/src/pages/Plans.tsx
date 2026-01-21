import { useState } from "react";
import { Card } from "../components/Card";
import { ProgressBar } from "../components/ProgressBar";

const periods = ["День", "Неделя", "Месяц", "Год"] as const;

export function Plans() {
  const [active, setActive] = useState<(typeof periods)[number]>("Месяц");

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

      <Card>
        <div className="section-header">
          <h3>Цели на {active.toLowerCase()}</h3>
          <span className="muted">3</span>
        </div>
        {[72, 48, 28].map((progress, index) => (
          <div key={index} className="goal-row">
            <div>
              <p>Цель #{index + 1}</p>
              <span className="muted">Прогресс</span>
            </div>
            <div className="goal-row__progress">
              <ProgressBar value={progress} />
              <span className="muted">{progress}%</span>
            </div>
          </div>
        ))}
      </Card>
    </div>
  );
}
