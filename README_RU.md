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
```

## Синхронизация (MVP v2)

- **Только pull** через `GET /sync?workspace_id=...&since=RFC3339`.
- Клиент сохраняет `lastSync = server_time` из ответа.
- `POST /sync` отключён; все изменения идут через CRUD.

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
- Environment: `VITE_API_URL` (URL backend на Render)

## Переменные окружения

Backend:
- `DATABASE_URL`
- `JWT_SECRET`
- `CORS_ORIGIN`
- `PORT`

Frontend:
- `VITE_API_URL`

## API заметки

- Все защищённые запросы требуют `Authorization: Bearer <token>`.
- Ошибки возвращаются в формате:
  ```json
  { "error": { "code": "INSUFFICIENT_FUNDS", "message": "Недостаточно огоньков" } }
  ```

## Документация

- [API EN](./docs/api_EN.md)
- [API RU](./docs/api_RU.md)
