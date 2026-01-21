import { themes, Theme } from "../theme/themes";
import { Card } from "../components/Card";

export function Settings({
  email,
  password,
  userEmail,
  workspaceId,
  theme,
  onEmailChange,
  onPasswordChange,
  onThemeChange,
  onLogin,
  onRegister,
  onSync,
  status,
  apiMissing,
  lastSync
}: {
  email: string;
  password: string;
  userEmail?: string | null;
  workspaceId?: string | null;
  theme: Theme;
  onEmailChange: (value: string) => void;
  onPasswordChange: (value: string) => void;
  onThemeChange: (theme: Theme) => void;
  onLogin: () => void;
  onRegister: () => void;
  onSync: () => void;
  status?: string | null;
  apiMissing: boolean;
  lastSync?: string | null;
}) {
  return (
    <div className="page">
      <header className="page__header">
        <h1>Настройки</h1>
        <p className="muted">Аккаунт, синхронизация и тема</p>
      </header>

      {apiMissing && <div className="alert">API URL не настроен. Проверьте VITE_API_BASE_URL.</div>}

      <Card>
        <div className="section-header">
          <h3>Профиль</h3>
          <span className="muted">{userEmail ?? "Гость"}</span>
        </div>
        <div className="muted">Workspace: {workspaceId ?? "—"}</div>
      </Card>

      <Card>
        <h3>Авторизация</h3>
        <div className="auth">
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(event) => onEmailChange(event.target.value)}
          />
          <input
            type="password"
            placeholder="Пароль"
            value={password}
            onChange={(event) => onPasswordChange(event.target.value)}
          />
          <div className="auth__actions">
            <button onClick={onLogin}>Войти</button>
            <button className="secondary" onClick={onRegister}>
              Регистрация
            </button>
          </div>
        </div>
        {status && <p className="muted">{status}</p>}
      </Card>

      <Card>
        <h3>Тема</h3>
        <div className="theme-grid">
          {themes.map((item) => (
            <button
              key={item.id}
              className={item.id === theme ? "active" : ""}
              onClick={() => onThemeChange(item.id)}
            >
              {item.label}
            </button>
          ))}
        </div>
      </Card>

      <Card>
        <h3>Синхронизация</h3>
        <p className="muted">Последняя: {lastSync ?? "—"}</p>
        <button className="full" onClick={onSync}>
          Sync now
        </button>
      </Card>
    </div>
  );
}
