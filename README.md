# Orencode Zenko

Единая цифровая платформа Федерации спортивного программирования Оренбургской области.

Этот README описывает полный запуск проекта через Docker Compose. Устанавливать Go, PostgreSQL и Node.js на компьютер для такого запуска не требуется.

## Состав Docker-окружения

| Сервис | Назначение | Адрес с компьютера |
|---|---|---|
| `postgres` | PostgreSQL и постоянное хранение данных | `127.0.0.1:${POSTGRES_PORT}` |
| `migrate` | Последовательное применение SQL-миграций | одноразовый контейнер |
| `rbac-seed` | Создание permissions, системных ролей и главного администратора | одноразовый контейнер |
| `seed-demo` | Добавление демонстрационных данных | одноразовый контейнер |
| `server` | Backend API на Go | `http://localhost:${BACKEND_PORT}` |
| `frontend` | Frontend на Next.js | `http://localhost:${FRONTEND_PORT}` |

При первом старте сервисы запускаются в следующем порядке:

```text
postgres → migrate → rbac-seed → seed-demo → server → frontend
```

Контейнеры `migrate`, `rbac-seed` и `seed-demo` после успешного выполнения завершаются с кодом `0`. Состояние `Exited (0)` для них является нормальным.

## Требования

Перед запуском установите:

- Git;
- Docker Desktop для Windows или macOS либо Docker Engine для Linux;
- Docker Compose v2, который вызывается командой `docker compose`.

Проверьте установку:

```bash
docker --version
docker compose version
```

Для сборки всех образов желательно иметь не менее 4 ГБ свободной оперативной памяти и несколько гигабайт свободного места на диске.

## 1. Получение проекта

```bash
git clone <URL_РЕПОЗИТОРИЯ>
cd orencode-zenko
```

Все следующие команды выполняются из корня проекта, где находится `docker-compose.yml`.

## 2. Настройка переменных окружения backend и Docker Compose

Основной файл окружения для полного Docker-запуска:

```text
backend/.env
```

Создайте его из примера.

```bash
cp backend/.env.example backend/.env
```

Откройте `backend/.env` и заполните значения. Рабочий шаблон для локального Docker-запуска:

```dotenv
# PostgreSQL
POSTGRES_USER=zenko
POSTGRES_PASSWORD=change_this_database_password
POSTGRES_DB=zenko
POSTGRES_PORT=5432

# Подключение с компьютера, например для migrate CLI или IDE
DATABASE_URL=postgres://zenko:change_this_database_password@localhost:5432/zenko?sslmode=disable

# Подключение контейнеров к PostgreSQL внутри Docker-сети
# Здесь хост всегда postgres, а внутренний порт всегда 5432
DATABASE_URL_DOCKER=postgres://zenko:change_this_database_password@postgres:5432/zenko?sslmode=disable

# Порты, открываемые на компьютере
BACKEND_PORT=8080
FRONTEND_PORT=3000

# JWT
JWT_SECRET=replace_with_a_long_random_secret_at_least_32_chars
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h
DEVICE_SESSION_TTL=12h

# Главный администратор, создаваемый rbac-seed
ADMIN_EMAIL=admin@example.local
ADMIN_PASSWORD=change_this_admin_password
ADMIN_FIRST_NAME=Главный
ADMIN_LAST_NAME=Администратор
ADMIN_BIRTH_DATE=1970-01-01

# Файлы и фоновые задачи
STORAGE_PATH=./private_uploads
STORAGE_CLEANUP_INTERVAL=10000h
STORAGE_CLEANUP_RETENTION=10000h
STORAGE_CLEANUP_BATCH=100
RANK_CLEANUP_INTERVAL=24h
DEADLINE_REMINDER_INTERVAL=1h

# Разрешённые источники браузерных запросов
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Необязательно. При пустом значении Telegram-уведомления отключены
TG_TOKEN_BOT=
```
## Вот заполненный на всякий случай

```dotenv
# PostgreSQL
POSTGRES_USER=postgres
POSTGRES_PASSWORD=StrongPassword123
POSTGRES_DB=fsp
POSTGRES_PORT=5432

# Подключение с компьютера (Windows)
DATABASE_URL=postgres://postgres:StrongPassword123@localhost:5432/fap?sslmode=disable

# Подключение контейнеров к PostgreSQL внутри Docker-сети
DATABASE_URL_DOCKER=postgres://postgres:StrongPassword123@postgres:5432/fsp?sslmode=disable

# Порты
BACKEND_PORT=8080
FRONTEND_PORT=3000

# JWT
JWT_SECRET=super_secret_key
JWT_ACCESS_TTL=15m
JWT_REFRESH_TTL=720h
DEVICE_SESSION_TTL=12h

# Главный администратор (RBAC seed)
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=admin12345!1
ADMIN_FIRST_NAME=Главный
ADMIN_LAST_NAME=Администратор
ADMIN_BIRTH_DATE=1970-01-01

# Файлы и фоновые задачи
STORAGE_PATH=./private_uploads
STORAGE_CLEANUP_INTERVAL=10000h
STORAGE_CLEANUP_RETENTION=10000h
STORAGE_CLEANUP_BATCH=100
RANK_CLEANUP_INTERVAL=24h
DEADLINE_REMINDER_INTERVAL=1h

# CORS
CORS_ALLOWED_ORIGINS=http://localhost:3000

# Telegram (пустое значение отключает уведомления)
TG_TOKEN_BOT=
```

### Правила заполнения `backend/.env`

#### PostgreSQL

`POSTGRES_USER`, `POSTGRES_PASSWORD` и `POSTGRES_DB` создают пользователя и базу при первом запуске пустого Docker volume.

Одинаковые логин, пароль и имя базы должны использоваться в `DATABASE_URL` и `DATABASE_URL_DOCKER`.

Различаются только адрес и порт:

- в `DATABASE_URL` используется `localhost` и значение `POSTGRES_PORT`;
- в `DATABASE_URL_DOCKER` всегда используются хост `postgres` и порт `5432`.

Например, если внешний порт изменён:

```dotenv
POSTGRES_PORT=5433
DATABASE_URL=postgres://zenko:password@localhost:5433/zenko?sslmode=disable
DATABASE_URL_DOCKER=postgres://zenko:password@postgres:5432/zenko?sslmode=disable
```

Если пароль содержит `@`, `:`, `/`, `#`, `%` или другие специальные символы URL, их нужно percent-encode в обеих строках подключения. Для первого локального запуска проще использовать длинный пароль из латинских букв, цифр, `_` и `-`.

#### JWT

`JWT_SECRET` должен быть длинной случайной строкой. После изменения секрета все ранее выданные access- и refresh-токены становятся недействительными.

Формат времени соответствует Go Duration:

- `15m` — 15 минут;
- `24h` — 24 часа;
- `720h` — 30 суток.

#### Главный администратор

Обязательно используется переменная `ADMIN_EMAIL`. Переменная `ADMIN_MAIL` backend не читает.

`ADMIN_PASSWORD` должен содержать от 12 до 72 байт. При старте `rbac-seed`:

- создаются или обновляются системные permissions;
- создаются роли `super_admin`, `admin`, `participant`, `coach` и `judge`;
- создаётся или активируется главный администратор;
- главному администратору назначаются системные административные роли.

Email и пароль из этих переменных используются для первого входа в приложение.

#### Хранилище файлов

В контейнере `server` значение `STORAGE_PATH` автоматически заменяется на:

```text
/app/private_uploads
```

Файлы сохраняются в Docker volume `uploaded_files` и не удаляются при обычном перезапуске контейнеров.

#### Telegram

`TG_TOKEN_BOT` необязателен. Оставьте его пустым, если Telegram-бот не используется.

Не добавляйте реальные `.env`-файлы и секреты в Git.

## 3. Настройка frontend env

Для полного запуска через Docker отдельный `frontend/.env` не требуется:

- frontend обращается к backend через относительные маршруты `/api/v1/*`;
- Next.js проксирует их на `http://server:8080` внутри Docker-сети;
- это значение уже передаётся в Dockerfile из `docker-compose.yml`.

Файл `frontend/.env` используется при локальном запуске frontend вне Docker. Его можно создать из примера:

```bash
cp frontend/.env.example frontend/.env
```

PowerShell:

```powershell
Copy-Item frontend/.env.example frontend/.env
```

Пример:

```dotenv
BACKEND_URL=http://localhost:8080
NEXT_PUBLIC_PERSONAL_CONSENT_TEMPLATE_URL=
NEXT_PUBLIC_PARENTAL_CONSENT_TEMPLATE_URL=
NEXT_PUBLIC_AFTER_AUTH_URL=/workspace
```

Назначение переменных:

| Переменная | Назначение |
|---|---|
| `BACKEND_URL` | Адрес backend для Next.js proxy при локальном запуске frontend |
| `NEXT_PUBLIC_PERSONAL_CONSENT_TEMPLATE_URL` | Необязательная ссылка на шаблон согласия на обработку персональных данных |
| `NEXT_PUBLIC_PARENTAL_CONSENT_TEMPLATE_URL` | Необязательная ссылка на шаблон родительского согласия |
| `NEXT_PUBLIC_AFTER_AUTH_URL` | Необязательный маршрут после входа, по умолчанию `/workspace` |

`frontend/.env` исключён из Docker build. Поэтому эти `NEXT_PUBLIC_*` значения относятся к локальной frontend-разработке. При пустых ссылках приложение продолжает работать, но ссылки на скачивание соответствующих шаблонов не отображаются.

## 4. Первый запуск backend, frontend и PostgreSQL

Запустите из корня проекта:

```bash
docker compose --env-file backend/.env up --build -d
```

Параметры команды:

- `--env-file backend/.env` подставляет переменные в `docker-compose.yml`;
- `--build` собирает актуальные backend- и frontend-образы;
- `-d` запускает контейнеры в фоне.

На первом запуске Docker скачает базовые образы, соберёт Go backend и Next.js frontend, создаст базу, применит миграции и выполнит seed. Это может занять несколько минут.

Не запускайте Compose без `--env-file backend/.env`, иначе `DATABASE_URL_DOCKER` не будет подставлен в команды миграций и seed.

## 5. Проверка запуска

Покажите состояние контейнеров:

```bash
docker compose --env-file backend/.env ps -a
```

Ожидаемое состояние:

- `postgres` — `Up` и `healthy`;
- `migrate` — `Exited (0)`;
- `rbac-seed` — `Exited (0)`;
- `seed-demo` — `Exited (0)`;
- `server` — `Up`;
- `frontend` — `Up`.

Проверьте логи:

```bash
docker compose --env-file backend/.env logs migrate rbac-seed seed-demo
docker compose --env-file backend/.env logs server frontend
```

Для просмотра логов в реальном времени:

```bash
docker compose --env-file backend/.env logs -f server frontend
```

После запуска доступны:

- frontend: [http://localhost:3000](http://localhost:3000);
- backend API: [http://localhost:8080](http://localhost:8080);
- публичный запрос для проверки backend: [http://localhost:8080/api/v1/competitions](http://localhost:8080/api/v1/competitions).

Если в `.env` изменены `FRONTEND_PORT` или `BACKEND_PORT`, используйте соответствующие порты вместо `3000` и `8080`.

Войдите через frontend с `ADMIN_EMAIL` и `ADMIN_PASSWORD`, указанными в `backend/.env`.

## 6. Обычный повторный запуск

Если исходный код и Dockerfile не менялись:

```bash
docker compose --env-file backend/.env up -d
```

Если код, зависимости или Dockerfile изменились:

```bash
docker compose --env-file backend/.env up --build -d
```

Для принудительной пересборки без Docker cache:

```bash
docker compose --env-file backend/.env build --no-cache server frontend
docker compose --env-file backend/.env up -d
```

## 7. Применение новых миграций и повторный seed

После получения изменений с новыми SQL-миграциями:

```bash
docker compose --env-file backend/.env build server
docker compose --env-file backend/.env run --rm migrate
docker compose --env-file backend/.env run --rm rbac-seed
docker compose --env-file backend/.env run --rm seed-demo
docker compose --env-file backend/.env up -d server frontend
```

Миграции применяются только вперёд и сохраняются в volume PostgreSQL. Seed-команды рассчитаны на повторный запуск и обновляют системные справочники и демонстрационные записи.

Для обычного обновления также можно выполнить одну команду:

```bash
docker compose --env-file backend/.env up --build -d
```

Отдельные команды выше удобнее при диагностике, потому что сразу видно, на каком этапе произошла ошибка.

## 8. Остановка и удаление контейнеров

Остановить проект, сохранив БД и загруженные файлы:

```bash
docker compose --env-file backend/.env down
```

Остановить без удаления контейнеров:

```bash
docker compose --env-file backend/.env stop
```

Повторно запустить остановленные контейнеры:

```bash
docker compose --env-file backend/.env start
```

## 9. Полный сброс локальных данных

Следующая команда удаляет контейнеры, базу PostgreSQL и загруженные файлы:

```bash
docker compose --env-file backend/.env down -v
```

После неё восстановить локальные данные нельзя. Используйте полный сброс только для тестовой БД.

Для создания чистого окружения после сброса:

```bash
docker compose --env-file backend/.env up --build -d
```

## 10. Запуск отдельных компонентов

Только PostgreSQL:

```bash
docker compose --env-file backend/.env up -d postgres
```

PostgreSQL, миграции, seed и backend без frontend:

```bash
docker compose --env-file backend/.env up --build -d postgres migrate rbac-seed seed-demo server
```

Перезапуск только backend:

```bash
docker compose --env-file backend/.env up --build -d server
```

Перезапуск только frontend:

```bash
docker compose --env-file backend/.env up --build -d frontend
```

## 11. Создание клиента внешней интеграции

Этот шаг не нужен для обычного запуска frontend и backend. Он используется только для подключения внешних систем.

Сначала соберите актуальный backend-образ, содержащий CLI-команду:

```bash
docker compose --env-file backend/.env build server
```

Затем создайте интеграционного клиента:

```bash
docker compose --env-file backend/.env run --rm \
  --entrypoint /seed-integration \
  -e INTEGRATION_CLIENT_ID=partner-system \
  -e INTEGRATION_CLIENT_SECRET=replace_with_strong_integration_secret \
  -e INTEGRATION_CLIENT_NAME="Partner system" \
  -e INTEGRATION_OWNER_EMAIL=admin@example.local \
  -e INTEGRATION_PERMISSIONS=integration.results.import,integration.webhooks.manage \
  server
```

Эквивалент для PowerShell:

```powershell
docker compose --env-file backend/.env run --rm --entrypoint /seed-integration -e INTEGRATION_CLIENT_ID=partner-system -e INTEGRATION_CLIENT_SECRET=replace_with_strong_integration_secret -e "INTEGRATION_CLIENT_NAME=Partner system" -e INTEGRATION_OWNER_EMAIL=admin@example.local -e INTEGRATION_PERMISSIONS=integration.results.import,integration.webhooks.manage server
```

Подробный контракт токенов, импорта результатов и webhooks находится в [документации внешних интеграций](./doc/integration-api.md).

## 12. Типовые проблемы

### `DATABASE_URL_DOCKER variable is not set`

Причина: Compose запущен без файла окружения.

Используйте:

```bash
docker compose --env-file backend/.env up --build -d
```

### `password authentication failed for user`

Проверьте совпадение `POSTGRES_USER`, `POSTGRES_PASSWORD` и `POSTGRES_DB` в обеих строках подключения.

Если PostgreSQL volume уже был создан с другими реквизитами, изменение `.env` не меняет пользователя внутри существующей базы. Для тестового окружения выполните полный сброс через `down -v` либо верните первоначальные реквизиты.

### Порт уже занят

Измените внешний порт в `backend/.env`:

```dotenv
POSTGRES_PORT=5433
BACKEND_PORT=8081
FRONTEND_PORT=3001
```

Не забудьте обновить `DATABASE_URL`, `CORS_ALLOWED_ORIGINS` и адреса, которые открываете в браузере. Внутренние адреса `postgres:5432` и `server:8080` менять не нужно.

### `migrate` завершился с ошибкой

Посмотрите лог:

```bash
docker compose --env-file backend/.env logs migrate
```

Не удаляйте и не переименовывайте уже применённые миграции в общей или production-базе.

### `rbac-seed` или `seed-demo` завершился с ошибкой

```bash
docker compose --env-file backend/.env logs rbac-seed seed-demo
```

Частые причины:

- миграции не применились;
- неверен `DATABASE_URL_DOCKER`;
- `ADMIN_PASSWORD` короче 12 символов;
- вместо `ADMIN_EMAIL` указано `ADMIN_MAIL`.

### Frontend открывается, но API недоступен

Проверьте, что backend запущен, и посмотрите его лог:

```bash
docker compose --env-file backend/.env ps server
docker compose --env-file backend/.env logs server
```

В Docker frontend должен использовать внутренний адрес `http://server:8080`, который уже задан в `docker-compose.yml`. Не заменяйте его на `localhost`: внутри frontend-контейнера `localhost` означает сам frontend-контейнер.

### Изменения frontend или backend не появились

Пересоберите нужный образ:

```bash
docker compose --env-file backend/.env up --build -d server frontend
```

Если Docker продолжает использовать старый слой:

```bash
docker compose --env-file backend/.env build --no-cache server frontend
docker compose --env-file backend/.env up -d
```

## 13. Полезные команды

```bash
# Состояние всех сервисов
docker compose --env-file backend/.env ps -a

# Все логи
docker compose --env-file backend/.env logs

# Последние 100 строк backend
docker compose --env-file backend/.env logs --tail=100 server

# Перезапуск backend
docker compose --env-file backend/.env restart server

# Открыть psql внутри контейнера
docker compose --env-file backend/.env exec postgres \
  sh -c 'psql -U "$POSTGRES_USER" -d "$POSTGRES_DB"'

# Выполнить backend-тесты в локальном Go-окружении
cd backend && go test ./...
```

Для команды `psql` в PowerShell значения можно указать явно из `backend/.env`, например:

```powershell
docker compose --env-file backend/.env exec postgres psql -U zenko -d zenko
```

## Документация

- [Общий каталог документации](./doc/README.md)
- [API для frontend](./doc/frontend-api.md)
- [API внешних интеграций](./doc/integration-api.md)
- [План проверки backend](./doc/plan_test.md)
- [Запуск Docker](./doc/docker.md)
