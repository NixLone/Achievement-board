import { useState } from "react";
import { Achievement } from "../storage";
import { Card } from "../components/Card";

export function Achievements({
  achievements,
  onAdd,
  onEdit,
  onSave,
  onDelete
}: {
  achievements: Achievement[];
  onAdd: (form: { title: string; description: string; imageUrl: string }) => void;
  onEdit: (achievement: Achievement, update: Partial<Achievement>) => void;
  onSave: (achievement: Achievement) => void;
  onDelete: (achievement: Achievement) => void;
}) {
  return (
    <div className="page">
      <header className="page__header">
        <h1>Достижения</h1>
        <p className="muted">Собирайте награды и опыт</p>
      </header>

      <Card>
        <AchievementForm onAdd={onAdd} />
        <div className="list">
          {achievements
            .filter((achievement) => !achievement.deleted_at)
            .map((achievement) => (
              <div key={achievement.id} className="list-item">
                <div className="list-item__icon">🏅</div>
                <div className="list-item__body">
                  <input
                    className="inline-input"
                    value={achievement.title}
                    onChange={(event) => onEdit(achievement, { title: event.target.value })}
                    onBlur={() => onSave(achievement)}
                  />
                  <input
                    className="inline-input muted"
                    value={achievement.description}
                    onChange={(event) => onEdit(achievement, { description: event.target.value })}
                    onBlur={() => onSave(achievement)}
                  />
                </div>
                <div className="list-item__trailing">
                  <button className="ghost" onClick={() => onDelete(achievement)}>
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

function AchievementForm({
  onAdd
}: {
  onAdd: (form: { title: string; description: string; imageUrl: string }) => void;
}) {
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [imageUrl, setImageUrl] = useState("");

  return (
    <div className="form-row">
      <input placeholder="Достижение" value={title} onChange={(event) => setTitle(event.target.value)} />
      <input placeholder="Описание" value={description} onChange={(event) => setDescription(event.target.value)} />
      <input placeholder="Image URL" value={imageUrl} onChange={(event) => setImageUrl(event.target.value)} />
      <button
        onClick={() => {
          if (!title) return;
          onAdd({ title, description, imageUrl });
          setTitle("");
          setDescription("");
          setImageUrl("");
        }}
      >
        Добавить
      </button>
    </div>
  );
}
