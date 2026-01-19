# FireGoals MVP

FireGoals — кроссплатформенный PWA для целей, задач, наград, достижений и общего баланса «огоньков» в рабочих пространствах.

## Стек

- Backend: Go + Postgres (Neon)
- Frontend: React + Vite + TypeScript (PWA)
- Деплой: Render (backend) + Cloudflare Pages (frontend)

## Быстрый старт

FireGoals состоит из двух компонентов — бэкенда на Go и фронтенда на React. Для локальной разработки удобнее всего использовать файл переменных окружения.

1. Установите необходимые зависимости:
   - **Go 1.22 или новее**
   - **Node.js 18 или новее**
   - **PostgreSQL 15 или новее**
2. Скопируйте файл `.env.example` в `.env` и заполните значения. Минимально нужно указать `DATABASE_URL`, `JWT_SECRET` и `VITE_API_URL`.
3. Создайте базу данных Postgres и примените миграции из каталога `migrations`.
4. Запустите бэкенд‑команду `go run ./cmd/server`.
5. В другом терминале запустите фронтенд: `cd frontend && npm install && npm run dev`.

Ниже приведены подробные инструкции для каждого этапа.

### Подготовка базы данных

Приложение использует PostgreSQL в качестве хранилища данных. Укажите строку подключения в переменной `DATABASE_URL` вида `postgresql://пользователь:пароль@хост:порт/имя_базы`.

Примените миграции, чтобы создать необходимые таблицы:

```bash
# пример применения миграции для локальной БД
psql "$DATABASE_URL" -f migrations/0001_init.sql
```

Если вы используете облачный сервис вроде Neon, создайте проект и базу, затем укажите полученную строку подключения в `.env`.

### Настройка backend

Бэкенд находится в папке `cmd/server`. Перед запуском убедитесь, что у вас скачаны все зависимости:

```bash
# загрузка зависимостей и обновление go.sum
go mod tidy

# чтение переменных из .env и запуск сервера
source .env
go run ./cmd/server
```

По умолчанию сервер будет слушать порт, указанный в `PORT` (например, `8080`). Для изменения используйте собственное значение в `.env`.

### Настройка frontend

Фронтенд расположен в каталоге `frontend` и использует Vite. После установки Node.js выполните:

```bash
cd frontend
# установка зависимостей (требует доступа к npm‑репозиторию)
npm install

# запуск dev‑сервера на http://localhost:5173
npm run dev
```

Фронтенд берёт адрес API из переменной `VITE_API_URL` (файл `.env`). Укажите туда URL бэкенда, например `http://localhost:8080`.

### Тестирование

В репозитории есть несколько unit‑тестов для слоя репозитория (`internal/repo`). Для их выполнения понадобится рабочая база данных Postgres. Установите переменную `DATABASE_URL` и запустите:

```bash
go test ./...
```

Если `DATABASE_URL` не задан, тесты будут пропущены. В процессе тестирования создаются временные схемы в базе, которые автоматически удаляются.

### Docker (опционально)

Для удобной локальной разработки можно воспользоваться Docker. Пример `Dockerfile` находится в корне проекта; его можно использовать вместе с вашей инфраструктурой (например, Render). Также вы можете настроить `docker-compose` с сервисами Postgres, backend и frontend. Пример файла не входит в репозиторий, но структура может выглядеть так:

```yaml
version: '3.9'
services:
  db:
    image: postgres:15
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: firegoals
    ports:
      - "5432:5432"
  backend:
    build: .
    environment:
      DATABASE_URL: postgresql://postgres:postgres@db:5432/firegoals
      JWT_SECRET: super-secret-string
      CORS_ORIGIN: http://localhost:5173
    depends_on:
      - db
    ports:
      - "8080:8080"
  frontend:
    build:
      context: ./frontend
    command: ["npm", "run", "dev"]
    environment:
      VITE_API_URL: http://localhost:8080
    ports:
      - "5173:5173"
    depends_on:
      - backend
```

### Дополнительные материалы

* Документацию по API вы найдёте в папке `docs` (русская и английская версии).
* Англоязычный вариант README находится в файле `README.md`.

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
