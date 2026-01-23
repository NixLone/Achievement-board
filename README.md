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
psql "$DATABASE_URL" -f migrations/0002_recurring_rewards_settings.sql
```

## Sync Model (MVP v2)

- **Pull-only sync** via `GET /sync?workspace_id=...&since=RFC3339`.
- Server captures a cursor time at the start of the request and returns `server_time = cursor_time`.
- Client stores `lastSync = server_time` from the response to avoid missing updates.
- `POST /sync` is disabled; CRUD endpoints are the source of writes.

### Client merge strategy

Merge entities by `id` and prefer the newest record by:
1) `updated_at` (server time), 2) `version` when timestamps match.

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
- Environment: `VITE_API_BASE_URL` (point to Render backend)

## Environment Variables

Backend:
- `DATABASE_URL`
- `JWT_SECRET`
- `CORS_ORIGIN`
- `PORT`

Frontend:
- `VITE_API_BASE_URL`

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

## Troubleshooting

**Render returns 404**  
Verify `/health` works and that the service is deployed from the repository root. Ensure `PORT` is configured by Render or set explicitly.

**Cloudflare Pages build fails on tsc**  
Install dependencies (`npm install`) and ensure `VITE_API_BASE_URL` is configured in Pages. Missing env can cause runtime errors.

**Cloudflare Pages root directory**  
Set the build root directory to `frontend/` and output directory to `dist`.

**DATABASE_URL typos**  
Make sure the connection string matches Neon format and includes the database name.

**Check /health**  
Open `<backend-url>/health` to verify the service is up.

**CORS/ENV issues**  
Make sure `CORS_ORIGIN` includes your Pages domain and `http://localhost:5173` for local dev.
