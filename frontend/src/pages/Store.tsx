import { useState } from "react";
import { Reward } from "../storage";
import { Card } from "../components/Card";

export function Store({
  rewards,
  onBuy,
  onAdd,
  onEdit,
  onSave,
  onDelete
}: {
  rewards: Reward[];
  onBuy: (reward: Reward) => void;
  onAdd: (form: { title: string; cost: number; description: string; icon: string }) => void;
  onEdit: (reward: Reward, update: Partial<Reward>) => void;
  onSave: (reward: Reward) => void;
  onDelete: (reward: Reward) => void;
}) {
  return (
    <div className="page">
      <header className="page__header">
        <h1>Магазин</h1>
        <p className="muted">Покупайте награды за огоньки</p>
      </header>

      <Card>
        <RewardForm onAdd={onAdd} />
        <div className="list">
          {rewards
            .filter((reward) => !reward.deleted_at)
            .map((reward) => (
              <div key={reward.id} className="list-item">
                <div className="list-item__icon">{reward.icon ?? "🎁"}</div>
                <div className="list-item__body">
                  <input
                    className="inline-input"
                    value={reward.title}
                    onChange={(event) => onEdit(reward, { title: event.target.value })}
                    onBlur={() => onSave(reward)}
                  />
                  <input
                    className="inline-input muted"
                    value={reward.description}
                    onChange={(event) => onEdit(reward, { description: event.target.value })}
                    onBlur={() => onSave(reward)}
                  />
                </div>
                <div className="list-item__trailing">
                  <div className="price-tag">{reward.cost} 🔥</div>
                  <button className="ghost" onClick={() => onBuy(reward)}>
                    Купить
                  </button>
                  <button className="ghost" onClick={() => onDelete(reward)}>
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

function RewardForm({
  onAdd
}: {
  onAdd: (form: { title: string; cost: number; description: string; icon: string }) => void;
}) {
  const [title, setTitle] = useState("");
  const [cost, setCost] = useState(20);
  const [description, setDescription] = useState("");
  const [icon, setIcon] = useState("🎁");

  return (
    <div className="form-row">
      <input placeholder="Награда" value={title} onChange={(event) => setTitle(event.target.value)} />
      <input type="number" value={cost} onChange={(event) => setCost(Number(event.target.value))} />
      <input placeholder="Описание" value={description} onChange={(event) => setDescription(event.target.value)} />
      <input placeholder="Emoji" value={icon} onChange={(event) => setIcon(event.target.value)} />
      <button
        onClick={() => {
          if (!title) return;
          onAdd({ title, cost, description, icon });
          setTitle("");
          setDescription("");
        }}
      >
        Добавить
      </button>
    </div>
  );
}
