import { useMemo, useState } from "react";
import { Reward, RewardPurchase } from "../storage";
import { Card } from "../components/Card";

export function Store({
  rewards,
  purchases,
  onBuy,
  onAdd,
  onEdit,
  onSave,
  onDelete
}: {
  rewards: Reward[];
  purchases: RewardPurchase[];
  onBuy: (reward: Reward) => void;
  onAdd: (form: { title: string; cost: number; description: string; icon: string; one_time: boolean }) => void;
  onEdit: (reward: Reward, update: Partial<Reward>) => void;
  onSave: (reward: Reward) => void;
  onDelete: (reward: Reward) => void;
}) {
  const purchasedRewardIds = useMemo(
    () => new Set(purchases.map((purchase) => purchase.reward_id)),
    [purchases]
  );
  const activeRewards = rewards.filter(
    (reward) => !reward.deleted_at && !(reward.one_time && purchasedRewardIds.has(reward.id))
  );
  const archivedPurchases = purchases.filter((purchase) => {
    const reward = rewards.find((item) => item.id === purchase.reward_id);
    return reward?.one_time;
  });

  return (
    <div className="page">
      <header className="page__header">
        <h1>Магазин</h1>
        <p className="muted">Покупайте награды за огоньки</p>
      </header>

      <Card>
        <RewardForm onAdd={onAdd} />
        <div className="list">
          {activeRewards.map((reward) => (
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
                {reward.one_time && <span className="tag">Одноразовая</span>}
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

      <Card>
        <div className="section-header">
          <h3>Архив покупок</h3>
          <span className="muted">{archivedPurchases.length}</span>
        </div>
        <div className="list">
          {archivedPurchases.map((purchase) => {
            const reward = rewards.find((item) => item.id === purchase.reward_id);
            return (
              <div key={purchase.id} className="list-item">
                <div className="list-item__icon">{reward?.icon ?? "✅"}</div>
                <div className="list-item__body">
                  <div>{reward?.title ?? "Награда"}</div>
                  <span className="muted">{new Date(purchase.purchased_at).toLocaleDateString()}</span>
                </div>
                <div className="list-item__trailing">
                  <div className="price-tag">{purchase.cost} 🔥</div>
                </div>
              </div>
            );
          })}
        </div>
      </Card>
    </div>
  );
}

function RewardForm({
  onAdd
}: {
  onAdd: (form: { title: string; cost: number; description: string; icon: string; one_time: boolean }) => void;
}) {
  const [title, setTitle] = useState("");
  const [cost, setCost] = useState(20);
  const [description, setDescription] = useState("");
  const [icon, setIcon] = useState("🎁");
  const [oneTime, setOneTime] = useState(false);

  return (
    <div className="form-row">
      <input placeholder="Награда" value={title} onChange={(event) => setTitle(event.target.value)} />
      <input type="number" value={cost} onChange={(event) => setCost(Number(event.target.value))} />
      <input placeholder="Описание" value={description} onChange={(event) => setDescription(event.target.value)} />
      <input placeholder="Emoji" value={icon} onChange={(event) => setIcon(event.target.value)} />
      <label className="checkbox">
        <input type="checkbox" checked={oneTime} onChange={(event) => setOneTime(event.target.checked)} />
        Одноразовая
      </label>
      <button
        onClick={() => {
          if (!title) return;
          onAdd({ title, cost, description, icon: icon || "🎁", one_time: oneTime });
          setTitle("");
          setDescription("");
          setIcon("🎁");
          setOneTime(false);
        }}
      >
        Добавить
      </button>
    </div>
  );
}
