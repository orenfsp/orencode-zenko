# Запуск проекта в Docker

Единый compose-файл находится в корне репозитория: `docker-compose.yml`.

## Подготовка

1. Создайте `backend/.env` из `backend/.env.example` и задайте значения переменных.
2. Для Docker значение `DATABASE_URL_DOCKER` должно использовать имя хоста `postgres`, например:

   ```env
   DATABASE_URL_DOCKER=postgres://postgres:postgres@postgres:5432/vera?sslmode=disable
   ```

3. Укажите `BACKEND_PORT=8080`. При необходимости задайте `FRONTEND_PORT=3000`.

## Запуск

Из корня репозитория:

```bash
docker compose --env-file backend/.env up --build -d
```

Compose запускает сервисы в следующем порядке:

```text
postgres → migrate → rbac-seed → seed-demo → server → frontend
```

После успешного запуска доступны:

- фронтенд: `http://localhost:3000`;
- backend API: `http://localhost:8080`;
- PostgreSQL: `127.0.0.1:${POSTGRES_PORT}`.

Фронтенд направляет запросы на `/api/v1/*` во внутренний backend-сервис, поэтому браузеру не требуется отдельный адрес API или настройка CORS.

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
