import { useState } from "react";
import { Card } from "../components/Card";

export function Workspace({
  workspaceId,
  members,
  onCreateWorkspace,
  onCreateInvite,
  onAcceptInvite,
  onRefresh
}: {
  workspaceId?: string | null;
  members: { id: string; email: string; role: string }[];
  onCreateWorkspace: (name: string) => void;
  onCreateInvite: () => void;
  onAcceptInvite: (code: string) => void;
  onRefresh: () => void;
}) {
  const [workspaceName, setWorkspaceName] = useState("");
  const [inviteCode, setInviteCode] = useState("");

  return (
    <div className="page">
      <header className="page__header">
        <h1>Общая цель</h1>
        <p className="muted">Управляйте общими целями</p>
      </header>

      <Card>
        <div className="section-header">
          <h3>Участники</h3>
          <button className="ghost" onClick={onRefresh}>
            Обновить
          </button>
        </div>
        <div className="list">
          {members.map((member) => (
            <div key={member.id} className="list-item">
              <div className="list-item__icon">👤</div>
              <div className="list-item__body">
                <div>{member.email}</div>
                <span className="muted">{member.role}</span>
              </div>
            </div>
          ))}
          {members.length === 0 && <p className="muted">Участников пока нет.</p>}
        </div>
      </Card>

      <Card>
        <div className="section-header">
          <h3>Приглашения</h3>
          <span className="muted">Workspace: {workspaceId ?? "—"}</span>
        </div>
        <div className="form-row">
          <button onClick={onCreateInvite}>Создать инвайт</button>
        </div>
        <div className="form-row">
          <input
            placeholder="Код приглашения"
            value={inviteCode}
            onChange={(event) => setInviteCode(event.target.value)}
          />
          <button
            onClick={() => {
              if (!inviteCode) return;
              onAcceptInvite(inviteCode);
              setInviteCode("");
            }}
          >
            Принять
          </button>
        </div>
      </Card>

      <Card>
        <h3>Новый workspace</h3>
        <div className="form-row">
          <input
            placeholder="Название"
            value={workspaceName}
            onChange={(event) => setWorkspaceName(event.target.value)}
          />
          <button
            onClick={() => {
              if (!workspaceName) return;
              onCreateWorkspace(workspaceName);
              setWorkspaceName("");
            }}
          >
            Создать
          </button>
        </div>
      </Card>
    </div>
  );
}
