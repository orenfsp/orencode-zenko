# Запуск проекта в Docker

Единый compose-файл находится в корне репозитория: `docker-compose.yml`.

## Подготовка

1. Создайте `backend/.env` из `backend/.env.example` и задайте значения переменных.
2. Для Docker значение `DATABASE_URL_DOCKER` должно использовать имя хоста `postgres`, например:

   ```env
   DATABASE_URL_DOCKER=postgres://postgres:postgres@postgres:5432/vera?sslmode=disable
   ```

3. Укажите `BACKEND_PORT=8080`. При необходимости задайте `FRONTEND_PORT=3000`.

Укажите также `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `JWT_SECRET`, `JWT_ACCESS_TTL`, `JWT_REFRESH_TTL` и `ADMIN_PASSWORD` (12–72 байта). Переменная `CORS_ALLOWED_ORIGINS` есть в конфигурации, но в текущем backend её значение не доходит до middleware; см. [ограничения CORS](./frontend-api.md#текущие-ограничения-контракта).

## Запуск

Из корня репозитория:

```bash
docker compose --env-file backend/.env up --build -d
```

Только backend с PostgreSQL и предусмотренными Compose миграциями и seed:

```bash
docker compose --env-file backend/.env build server
docker compose --env-file backend/.env up -d postgres migrate rbac-seed seed-demo server
```

Порты хоста можно изменить переменными `POSTGRES_PORT` и `BACKEND_PORT`; внутри сети Compose PostgreSQL остаётся на `postgres:5432`, backend — на `server:8080`.

Compose запускает сервисы в следующем порядке:

```text
postgres → migrate → rbac-seed → seed-demo → server → frontend
```

После успешного запуска доступны:

- фронтенд: `http://localhost:3000`;
- backend API: `http://localhost:8080`;
- PostgreSQL: `127.0.0.1:${POSTGRES_PORT}`.

При использовании frontend-прокси запросы на `/api/v1/*` направляются во внутренний backend-сервис. Для прямых запросов браузера с другого origin действуют указанные ограничения CORS.

## Полезные команды

```bash
# Посмотреть состояние и логи всех сервисов
docker compose --env-file backend/.env ps
docker compose --env-file backend/.env logs -f

# Остановить контейнеры, оставив данные PostgreSQL и загруженные файлы
docker compose --env-file backend/.env down

# Полностью очистить локальные данные проекта
docker compose --env-file backend/.env down -v
```
