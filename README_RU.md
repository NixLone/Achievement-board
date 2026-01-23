# FireGoals MVP

FireGoals — кроссплатформенный PWA для целей, задач, наград, достижений и общего баланса «огоньков» в рабочих пространствах.

## Стек

- Backend: Go + Postgres (Neon)
- Frontend: React + Vite + TypeScript (PWA)
- Деплой: Render (backend) + Cloudflare Pages (frontend)

## Запуск локально

### Требования

- Go 1.22+
- Node.js 18+
- Postgres 15+

### Backend

```bash
export DATABASE_URL=postgresql://user:pass@host/db
export JWT_SECRET=your-secret
export CORS_ORIGIN=http://localhost:5173
export PORT=8080

go run ./cmd/server
```

### Frontend

```bash
cd frontend
npm install
npm run dev
```

## База данных (Neon)

1. Создайте проект и БД в Neon.
2. Возьмите строку подключения и задайте `DATABASE_URL`.
3. Выполните миграции:

```bash
psql "$DATABASE_URL" -f migrations/0001_init.sql
psql "$DATABASE_URL" -f migrations/0002_recurring_rewards_settings.sql
```

## Синхронизация (MVP v2)

- **Только pull** через `GET /sync?workspace_id=...&since=RFC3339`.
- Сервер фиксирует cursor time в начале запроса и возвращает `server_time = cursor_time`.
- Клиент сохраняет `lastSync = server_time` из ответа, чтобы не терять изменения.
- `POST /sync` отключён; все изменения идут через CRUD.

### Стратегия мерджа на клиенте

Мердж по `id` с выбором более свежей записи по:
1) `updated_at` (время сервера), 2) `version` при равных timestamp.

## Гарантии экономики

- Выполнение задачи идемпотентно: повторные запросы не начисляют огоньки.
- Покупка награды невозможна при недостатке баланса.
- Операции task complete / reward buy / accept invite атомарны.

## Деплой

### Render (Backend)

1. Создайте Web Service.
2. Используйте Docker build.
3. Укажите переменные окружения:
   - `DATABASE_URL`
   - `JWT_SECRET`
   - `CORS_ORIGIN`
   - `PORT` (опционально, Render подставит автоматически)
4. Health check: `/health`.

### Cloudflare Pages (Frontend)

- Build command: `npm run build`
- Output directory: `dist`
- Environment: `VITE_API_BASE_URL` (URL backend на Render)

## Переменные окружения

Backend:
- `DATABASE_URL`
- `JWT_SECRET`
- `CORS_ORIGIN`
- `PORT`

Frontend:
- `VITE_API_BASE_URL`

## API заметки

- Все защищённые запросы требуют `Authorization: Bearer <token>`.
- Ошибки возвращаются в формате:
  ```json
  { "error": { "code": "INSUFFICIENT_FUNDS", "message": "Недостаточно огоньков" } }
  ```

## Документация

- [API EN](./docs/api_EN.md)
- [API RU](./docs/api_RU.md)

## Troubleshooting

**Render показывает 404**  
Проверьте `/health` и убедитесь, что сервис деплоится из корня репозитория. Убедитесь, что `PORT` корректно задан Render.

**Cloudflare Pages падает на tsc**  
Установите зависимости (`npm install`) и проверьте, что `VITE_API_BASE_URL` задан в Pages.

**Корневая директория Cloudflare Pages**  
Укажите root directory `frontend/` и output directory `dist`.

**Опечатки в DATABASE_URL**  
Проверьте формат строки подключения Neon и имя базы данных.

**Как проверить /health**  
Откройте `<backend-url>/health` и убедитесь, что статус `ok`.

**CORS/ENV**  
`CORS_ORIGIN` должен включать домен Pages и `http://localhost:5173` для локальной разработки.
