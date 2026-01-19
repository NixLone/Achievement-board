# FireGoals MVP

FireGoals is a cross-platform PWA for goals, tasks, fire currency, rewards, and achievements with workspace collaboration.

## Stack

- Backend: Go + Postgres (Neon)
- Frontend: React + Vite + TypeScript (PWA)
- Deploy: Render (backend) + Cloudflare Pages (frontend)

## Local Development

### Backend

```bash
export DATABASE_URL=postgresql://user:pass@host/db
export JWT_SECRET=your-secret
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

## Deployment

### Render (Backend)

1. Create a new Web Service.
2. Use Docker build.
3. Set environment variables:
   - `DATABASE_URL`
   - `JWT_SECRET`
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
- `PORT`

Frontend:
- `VITE_API_URL`

## API Notes

- All authenticated requests require `Authorization: Bearer <token>`.
- Sync endpoints:
  - `GET /sync?workspace_id=...&since=RFC3339`
  - `POST /sync` with `{ workspace_id, changes }`
