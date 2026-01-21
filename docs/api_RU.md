# FireGoals API (RU)

## Авторизация

### POST /auth/register

Запрос:
```json
{ "email": "user@example.com", "password": "secret" }
```

Ответ:
```json
{ "id": "<user-id>" }
```

### POST /auth/login

Запрос:
```json
{ "email": "user@example.com", "password": "secret" }
```

Ответ:
```json
{ "access_token": "<jwt>", "refresh_token": "<token>" }
```

### GET /me

Ответ:
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

Ответ complete:
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

## Sync (только pull)

### GET /sync

Параметры:
- `workspace_id`
- `since` (RFC3339 timestamp из предыдущего ответа)

Ответ:
```json
{ "changes": { "tasks": [], "rewards": [] }, "server_time": "2024-01-01T00:00:00Z" }
```

Примечания:
- `server_time` фиксируется в начале запроса (cursor time).
- Клиент должен сохранять `lastSync = server_time`.
- `POST /sync` отключён в MVP v2.

## Ошибки

Все ошибки возвращаются в формате:
```json
{ "error": { "code": "INSUFFICIENT_FUNDS", "message": "Недостаточно огоньков" } }
```

Коды ошибок:
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

## Клиентский merge-by-id

При синхронизации клиент должен мерджить сущности по `id` и оставлять более свежую запись:
- приоритет по `updated_at`
- при равенстве — по `version`
