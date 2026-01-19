# FireGoals MVP

FireGoals is a cross-platform PWA for goals, tasks, rewards, achievements, and a shared fire-currency balance across workspaces.

## Stack

- Backend: Go + Postgres (Neon)
- Frontend: React + Vite + TypeScript (PWA)
- Deploy: Render (backend) + Cloudflare Pages (frontend)

## Local Development

### Requirements

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

## Database Setup (Neon)

1. Create a Neon project and database.
2. Copy the connection string and set `DATABASE_URL`.
3. Run migrations:

```bash
psql "$DATABASE_URL" -f migrations/0001_init.sql
```

## Sync Model (MVP v2)

- **Pull-only sync** via `GET /sync?workspace_id=...&since=RFC3339`.
- Client stores `lastSync = server_time` from the response.
- `POST /sync` is disabled; CRUD endpoints are the source of writes.

## Economy Guarantees

- Task completion is idempotent: repeated calls do not earn more.
- Reward purchases are blocked if balance is insufficient.
- Task complete / reward buy / invite accept are atomic DB transactions.

## Deployment

### Render (Backend)

1. Create a new Web Service.
2. Use Docker build.
3. Set environment variables:
   - `DATABASE_URL`
   - `JWT_SECRET`
   - `CORS_ORIGIN`
   - `PORT` (optional; Render sets it automatically)
4. Health check: `/health`.

### Cloudflare Pages (Frontend)

- Build command: `npm run build`
- Output directory: `dist`
- Environment: `VITE_API_URL` (point to Render backend)

## Environment Variables

Backend:
- `DATABASE_URL`
- `JWT_SECRET`
- `CORS_ORIGIN`
- `PORT`

Frontend:
- `VITE_API_URL`

## API Notes

- Authenticated requests require `Authorization: Bearer <token>`.
- Errors return:
  ```json
  { "error": { "code": "INSUFFICIENT_FUNDS", "message": "Недостаточно огоньков" } }
  ```

## Docs

- [Russian README](./README_RU.md)
- [API EN](./docs/api_EN.md)
- [API RU](./docs/api_RU.md)
