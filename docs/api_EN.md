# FireGoals API (EN)

## Auth

### POST /auth/register

Request:
```json
{ "email": "user@example.com", "password": "secret" }
```

Response:
```json
{ "id": "<user-id>" }
```

### POST /auth/login

Request:
```json
{ "email": "user@example.com", "password": "secret" }
```

Response:
```json
{ "access_token": "<jwt>", "refresh_token": "<token>" }
```

### GET /me

Response:
```json
{ "id": "<user-id>", "email": "user@example.com" }
```

## Workspaces

- `GET /workspaces`
- `POST /workspaces`
- `GET /workspaces/{id}/balance`
- `POST /workspaces/{id}/invite`
- `POST /invites/accept`

## Goals

- `GET /goals?workspace_id=...`
- `POST /goals`
- `PUT /goals/{id}`
- `DELETE /goals/{id}?workspace_id=...`

## Tasks

- `GET /tasks?workspace_id=...`
- `POST /tasks`
- `PUT /tasks/{id}`
- `DELETE /tasks/{id}?workspace_id=...`
- `POST /tasks/{id}/complete`

Complete response:
```json
{ "earned": 10, "completed": true }
```

## Rewards

- `GET /rewards?workspace_id=...`
- `POST /rewards`
- `PUT /rewards/{id}`
- `DELETE /rewards/{id}?workspace_id=...`
- `POST /rewards/{id}/buy`

## Achievements

- `GET /achievements?workspace_id=...`
- `POST /achievements`
- `PUT /achievements/{id}`
- `DELETE /achievements/{id}?workspace_id=...`

## Sync (Pull-only)

### GET /sync

Query params:
- `workspace_id`
- `since` (RFC3339 timestamp from previous response)

Response:
```json
{ "changes": { "tasks": [], "rewards": [] }, "server_time": "2024-01-01T00:00:00Z" }
```

Notes:
- `server_time` is captured at request start (cursor time).
- Client must save `lastSync = server_time`.
- `POST /sync` is disabled in MVP v2.

## Errors

All errors return:
```json
{ "error": { "code": "INSUFFICIENT_FUNDS", "message": "Недостаточно огоньков" } }
```

Common codes:
- `INVALID_JSON`
- `VALIDATION_ERROR`
- `UNAUTHORIZED`
- `TOKEN_EXPIRED`
- `FORBIDDEN`
- `NOT_FOUND`
- `INSUFFICIENT_FUNDS`
- `INVITE_EXPIRED`
- `INVITE_USED`
- `SYNC_PUSH_DISABLED`
- `INTERNAL_ERROR`

## Client merge-by-id

When syncing, the client should merge entities by `id` and keep the newest record:
- prefer higher `updated_at`
- if equal, prefer higher `version`
