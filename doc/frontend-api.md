# API для фронтенда: Auth, профиль и RBAC

Документ описывает фактически реализованный контракт backend на текущий момент.

## Общие правила

Базовый префикс API:

```text
/api/v1
```

Для JSON-запросов необходимо отправлять:

```http
Content-Type: application/json
```

Тело обычного JSON-запроса ограничено 1 МБ. Неизвестные JSON-поля считаются ошибкой и приводят к `400 Bad Request`. Регистрация является исключением: она принимает `multipart/form-data` общим размером до 22 МиБ.

Успешный ответ с данными имеет форму:

```json
{
  "data": {}
}
```

Ответ с ошибкой:

```json
{
  "message": "Описание ошибки"
}
```

Для защищённых методов access-токен передаётся в заголовке:

```http
Authorization: Bearer <access_token>
```

`401 Unauthorized` означает, что токен или сессия отсутствуют, недействительны либо истекли. `403 Forbidden` означает, что пользователь авторизован, но у его ролей нет необходимого permission.

## Токены и сессия

Backend выдаёт два JWT:

- `access_token` — используется для запросов к защищённым endpoint’ам;
- `refresh_token` — используется для продления сессии и ротации токенов.

Срок жизни задаётся переменными backend `JWT_ACCESS_TTL` и `JWT_REFRESH_TTL`. Текущие значения из примера окружения: 15 минут и 720 часов.

Основные JWT claims:

```json
{
  "user_id": 42,
  "role_id": 7,
  "principal_type": "user",
  "type": "access",
  "exp": 1760000000,
  "iat": 1759999100,
  "jti": "..."
}
```

Сразу после обычной регистрации у пользователя нет роли, поэтому `role_id` в JWT отсутствует. После выбора роли backend возвращает новую пару токенов с `role_id`.

Frontend обязан заменять одновременно оба токена после успешного refresh или выбора роли. Старый refresh-токен после ротации больше не действует.

## Регистрация

```http
POST /api/v1/auth/register
```

Авторизация не требуется.

### Запрос

Запрос отправляется одним `multipart/form-data`. Frontend не загружает документы заранее и не передаёт URL или внутренние ID файлов. При использовании браузерного `FormData` заголовок `Content-Type` вручную задавать нельзя: браузер сам добавит корректный `boundary`.

Поля:

| Поле | Обязательно | Правила |
|---|---:|---|
| `first_name` | да | 1–100 символов |
| `last_name` | да | 1–100 символов |
| `middle_name` | нет | до 100 символов |
| `birth_date` | да | формат `YYYY-MM-DD`, не в будущем, возраст не более 120 лет |
| `email` | да | корректный email, до 255 символов; backend приводит к нижнему регистру |
| `password` | да | от 8 до 72 байт |
| `personal_data_consent_file` | да | файл согласия на обработку персональных данных |
| `parental_consent_document_file` | условно | письменное разрешение родителей; обязательно только для пользователя младше 18 лет |

Возраст рассчитывается backend по полной дате рождения. Пользователь считается совершеннолетним непосредственно в день 18-летия.

Для обоих файлов разрешены PDF, JPEG, PNG, XML и DOCX. Максимальный размер каждого файла — 10 МиБ. Backend проверяет фактическое содержимое, а не только имя или присланный MIME. Пустые файлы и ZIP-архивы, которые не являются корректной структурой DOCX, отклоняются.

Пример для браузера:

```ts
const form = new FormData();
form.append("first_name", values.firstName);
form.append("last_name", values.lastName);
if (values.middleName) form.append("middle_name", values.middleName);
form.append("birth_date", values.birthDate);
form.append("email", values.email);
form.append("password", values.password);
form.append("personal_data_consent_file", values.personalConsentFile);

// Добавлять только если пользователю ещё нет 18 лет.
if (values.parentalConsentFile) {
  form.append("parental_consent_document_file", values.parentalConsentFile);
}

await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
  method: "POST",
  body: form,
});
```

Пример регистрации несовершеннолетнего через curl:

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -F "first_name=Иван" \
  -F "last_name=Иванов" \
  -F "birth_date=2012-05-10" \
  -F "email=minor@example.com" \
  -F "password=strong-password" \
  -F "personal_data_consent_file=@./personal-consent.pdf" \
  -F "parental_consent_document_file=@./parental-consent.docx"
```

Для совершеннолетнего поле `parental_consent_document_file` передавать не нужно: лишний родительский документ будет отклонён с `400`.

Сначала backend проверяет текстовые поля и возраст, затем сохраняет файлы, а после этого одной транзакцией создаёт пользователя, записи согласий и привязывает файлы. Если регистрация не удалась, незакреплённые файлы удаляются компенсацией.

### Успешный ответ

Статус `201 Created`:

```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "..."
  }
}
```

Первый токен обычного пользователя не содержит `role_id`.

Возможные ошибки:

- `400` — некорректные/повторяющиеся/неизвестные поля, отсутствует обязательный файл или состав файлов не соответствует возрасту;
- `409` — email уже зарегистрирован;
- `413` — отдельный файл больше 10 МиБ либо вся multipart-форма больше 22 МиБ;
- `415` — содержимое файла имеет неподдерживаемый формат;
- `500` — внутренняя ошибка.

## Вход

```http
POST /api/v1/auth/login
```

### Запрос

```json
{
  "email": "user@example.com",
  "password": "strong-password"
}
```

### Ответ

Статус `200 OK`:

```json
{
  "data": {
    "access_token": "...",
    "refresh_token": "..."
  }
}
```

Ошибки:

- `401` — неверный email/пароль либо доступ пользователя отключён;
- `500` — внутренняя ошибка.

## Обновление токенов

```http
POST /api/v1/auth/refresh
```

### Запрос

```json
{
  "refresh_token": "..."
}
```

### Ответ

Статус `200 OK`:

```json
{
  "data": {
    "access_token": "new-access-token",
    "refresh_token": "new-refresh-token"
  }
}
```

После ответа необходимо удалить старую пару токенов и сохранить новую. Желательно использовать на фронте единый refresh-lock: при нескольких одновременных `401` выполнять только один refresh, а остальные запросы ставить в очередь.

Ошибки:

- `400` — `refresh_token` не передан;
- `401` — refresh-токен или сессия недействительны/истекли.

## Выход

```http
POST /api/v1/auth/logout
```

### Запрос

```json
{
  "refresh_token": "..."
}
```

Успешный ответ: `204 No Content` без тела. После ответа фронт удаляет оба токена и локальные данные текущего пользователя.

## Получение профиля

```http
GET /api/v1/profile
Authorization: Bearer <access_token>
```

Ответ содержит рабочие данные текущего пользователя. `password_hash`, `created_at` и `updated_at` в запрос не выбираются и в ответ не попадают.

```json
{
  "data": {
    "id": 42,
    "first_name": "Иван",
    "last_name": "Иванов",
    "middle_name": "Иванович",
    "birth_date": "2000-05-10",
    "email": "user@example.com",
    "tg_username": "ivanov",
    "access_status": true,
    "municipality": { "id": 3, "name": "Оренбург" },
    "organization": { "id": 8, "name": "Спортивная школа" },
    "roles": [
      { "id": 2, "code": "participant", "name": "Участник" }
    ]
  }
}
```

Если муниципалитет или организация не выбраны, соответствующее поле будет `null` или отсутствует. При отсутствии ролей возвращается пустой массив `roles`.

## Изменение профиля и выбор роли

```http
PATCH /api/v1/profile
Authorization: Bearer <access_token>
```

Метод принимает хотя бы одно из полей `tg_username` или `role_id`.

### Установка Telegram username

```json
{
  "tg_username": "ivan_ivanov"
}
```

Допускается значение с начальным `@`, backend его удалит. После нормализации username должен содержать 5–32 символа: латинские буквы, цифры и `_`.

Для удаления username отправляется пустая строка:

```json
{
  "tg_username": ""
}
```

Ответ:

```json
{
  "data": {
    "tg_username": "ivan_ivanov"
  }
}
```

### Выбор ролей

```json
{
  "role_ids": [7, 8],
  "refresh_token": "..."
}
```

Пользователь может самостоятельно выбрать одну или несколько ролей с `self_assignable: true`. Seed создаёт доступные для выбора роли `participant` («Участник»), `coach` («Тренер») и `judge` («Судья»). Системную роль главного администратора выбрать нельзя. Первый элемент `role_ids` становится активной ролью в новой паре JWT. Для обратной совместимости одиночное поле `role_id` также поддерживается.

Текущая реализация атомарно заменяет прежний набор самостоятельно выбранных ролей и ротирует токены в той же сессии. Выход из системы не нужен: frontend получает новую пару токенов и сразу заменяет ею старую.

```json
{
  "data": {
    "role_id": 7,
    "role_ids": [7, 8],
    "access_token": "new-access-token",
    "refresh_token": "new-refresh-token"
  }
}
```

Frontend должен сразу сохранить новую пару токенов. Если `refresh_token` не передан, роли не будут изменены.

Можно одновременно изменить Telegram username и выбрать роль:

```json
{
  "tg_username": "ivan_ivanov",
  "role_ids": [7, 8],
  "refresh_token": "..."
}
```

Ошибки:

- `400` — пустой PATCH, неверный username, роль недоступна для самостоятельного выбора или не передан refresh-токен;
- `401` — access/refresh-токен недействителен;
- `409` — Telegram username уже занят.

## Скачивание документа

```http
GET /api/v1/files/{fileID}/download
Authorization: Bearer <access_token>
```

Endpoint возвращает бинарное содержимое файла, а не JSON. Скачать документ может:

- пользователь-владелец документа;
- пользователь с permission `files.read` (в seed он есть у главного администратора).

Основные заголовки ответа:

```http
Content-Type: application/pdf
Content-Disposition: attachment; filename="consent.pdf"
Content-Length: 12345
Cache-Control: no-store
X-File-Path: personal_data_consent/<storage-key>.pdf
```

`X-File-Path` содержит внутренний относительный путь файла. Для скачивания во frontend удобно получать ответ как `Blob` и брать имя из `Content-Disposition`. Публичной раздачи каталога с файлами нет.

Ошибки:

- `400` — неверный `fileID`;
- `401` — отсутствует или недействителен access-токен;
- `403` — документ принадлежит другому пользователю и нет `files.read`;
- `404` — файл не существует или ещё не прикреплён к регистрации;
- `500` — ошибка БД или файлового хранилища.

### Список загруженных файлов

```http
GET /api/v1/admin/files?limit=50&offset=0
Authorization: Bearer <access_token>
```

Требуется `files.read`. Метод возвращает метаданные всех файлов из `uploaded_file`, включая статусы `uploaded`, `attached` и `delete_pending`; содержимое файлов не передаётся. Поддерживаются `limit` от 1 до 100 (по умолчанию 50) и неотрицательный `offset`.

```json
{
  "data": [
    {
      "id": 21,
      "purpose": "personal_data_consent",
      "storage_key": "personal_data_consent/4d2f...pdf",
      "original_filename": "consent.pdf",
      "content_type": "application/pdf",
      "size_bytes": 12345,
      "sha256": "...",
      "status": "attached",
      "created_at": "2026-10-27T09:30:00Z",
      "updated_at": "2026-10-27T09:30:00Z"
    }
  ]
}
```

Для получения содержимого используйте отдельный `GET /api/v1/files/{fileID}/download`.

## Заявки на спортивный разряд

### Подача заявки

```http
POST /api/v1/rank-applications
Authorization: Bearer <access_token>
Content-Type: application/json
```

Спортсмен без подтверждённого тренера по дисциплине подаёт заявку сам и не передаёт `athlete_user_id`:

```json
{
  "rank_id": 4,
  "discipline_id": 2
}
```

Подтверждённый тренер подаёт заявку за своего спортсмена:

```json
{
  "athlete_user_id": 42,
  "rank_id": 4,
  "discipline_id": 2
}
```

Если у спортсмена есть подтверждённый тренер по этой дисциплине, самостоятельная подача запрещена. Тренер не может подать заявку за спортсмена, с которым нет подтверждённой связи. Одновременно может существовать только одна заявка `pending` на одинаковые разряд и дисциплину.

Успешный ответ: `201 Created`.

### Мои заявки

```http
GET /api/v1/rank-applications/mine
Authorization: Bearer <access_token>
```

Ответ `200 OK`:

```json
{
  "data": [
    {
      "id": 17,
      "athlete_user_id": 42,
      "rank_id": 4,
      "discipline_id": 2,
      "submitted_by_user_id": 8,
      "status": "pending",
      "created_at": "2026-10-27T09:30:00Z"
    }
  ]
}
```

### Рассмотрение заявки

```http
PATCH /api/v1/admin/rank-applications/{applicationID}
Authorization: Bearer <access_token>
Content-Type: application/json
```

Требуется существующий административный permission `roles.assign`.

```json
{
  "status": "approved"
}
```

или:

```json
{
  "status": "rejected",
  "rejection_reason": "Не приложены необходимые подтверждающие сведения"
}
```

Допустимы только `approved` и `rejected`; отклонённая заявка требует причину. При одобрении PostgreSQL-триггер присваивает разряд спортсмену. При одобрении или отклонении он также добавляет уведомление спортсмену в `notification_outbox`.

Если в записи результата соревнования установлен `awarded_rank_id` и результат имеет статус `finished`, другой триггер автоматически присваивает этот разряд участнику соревнования.

Срок действия разряда задаётся полем `ttl` справочника разрядов в календарных годах. Backend удаляет истёкшие записи при старте и далее раз в `RANK_CLEANUP_INTERVAL` (по умолчанию — раз в сутки). `ttl = null` означает бессрочный разряд.

## Дашборд федерации

```http
GET /api/v1/admin/dashboard?year=2026
Authorization: Bearer <access_token>
```

Требуется permission `dashboard.read`. Параметр `year` необязательный: если его нет, backend использует текущий календарный год. Допустимый диапазон года — от 2000 до следующего календарного года.

Метрики считаются только по данным PostgreSQL и учитывают пользователей с `access_status = true`:

- `athletes` — активные пользователи с ролью `participant`;
- `competitions` — соревнования выбранного года, кроме отменённых;
- `participants` — уникальные спортсмены с подтверждёнными заявками на соревнования выбранного года;
- `active_ranks` — действующие записи присвоенных разрядов.

Ответ содержит также списки ближайших соревнований, заявок на модерацию, истекающих разрядов и судейских категорий, а также регистрационных и отчётных дедлайнов. При отсутствии данных backend возвращает пустые массивы, а не `null`.

Пример ответа:

```json
{
  "data": {
    "period_year": 2026,
    "metrics": {
      "athletes": 120,
      "competitions": 18,
      "participants": 94,
      "active_ranks": 76
    },
    "upcoming_competitions": [],
    "pending_competition_applications": [],
    "pending_rank_applications": [],
    "pending_judge_category_applications": [],
    "expiring_ranks": [],
    "expiring_judge_categories": [],
    "deadlines": []
  }
}
```

Ошибки: `400` для некорректного `year`, `401` без access-токена, `403` без `dashboard.read`, `500` при ошибке базы данных.

## Единый реестр

### Пользователи

```http
GET /api/v1/admin/registry/users
Authorization: Bearer <access_token>
```

Требуется `registry.read`. Параметры фильтрации обрабатываются backend и комбинируются:

| Параметр | Описание |
|---|---|
| `search` | Поиск по нормализованному ФИО. |
| `role` | Код роли: `participant`, `coach` или `judge`. |
| `discipline_id` | Дисциплина пользователя. |
| `organization_id` | Организация пользователя. |
| `rank_id` | Действующий разряд. |
| `age_group_id` | Текущая возрастная группа. |
| `municipality_id` | Муниципалитет пользователя. |
| `limit` | От 1 до 100, по умолчанию 50. |
| `offset` | Неотрицательное смещение, по умолчанию 0. |

В элементе пользователя возвращаются роли, дисциплины, действующие разряды, действующая судейская категория, а также:

- `coach_relations` — все связи тренера и спортсмена с типом связи (`coach` или `participant`), дисциплиной, статусом и датами;
- `judge_assignments` — назначения судьи с соревнованием, должностью, статусом соревнования и датами.

Исторические завершённые связи и назначения не скрываются из этих вложенных списков. При отсутствии связей возвращаются пустые массивы.

Пример элемента:

```json
{
  "id": 7,
  "last_name": "Иванов",
  "first_name": "Иван",
  "coach_relations": [
    {
      "id": 12,
      "relation_type": "participant",
      "related_user_id": 31,
      "related_user_name": "Петров Пётр",
      "discipline": { "id": 2, "name": "Спортивное программирование" },
      "status": "confirmed",
      "requested_at": "2026-01-10T10:00:00Z",
      "confirmed_at": "2026-01-11T10:00:00Z"
    }
  ],
  "judge_assignments": [
    {
      "id": 4,
      "competition_id": 9,
      "competition_name": "Региональный чемпионат",
      "competition_status": "published",
      "position_id": 1,
      "position_name": "Главный судья",
      "start_at": "2026-05-20T09:00:00Z",
      "end_at": "2026-05-21T18:00:00Z",
      "appointed_at": "2026-04-01T12:00:00Z"
    }
  ]
}
```

Ответ имеет формат страницы:

```json
{ "data": { "items": [], "total": 0, "limit": 50, "offset": 0 } }
```

### Организации

```http
GET /api/v1/admin/registry/organizations?search=школа&municipality_id=3&limit=50&offset=0
Authorization: Bearer <access_token>
```

Требуется `registry.read`. Организация возвращается с названием, типом и объектом муниципалитета. При отсутствии записей `items` — пустой массив.

## Permissions

Сейчас определены следующие permissions:

| Permission | Назначение |
|---|---|
| `roles.read` | просмотр ролей |
| `roles.create` | создание ролей |
| `roles.assign` | назначение ролей пользователям |
| `permissions.read` | просмотр справочника permissions |
| `files.read` | просмотр документов любых пользователей |
| `dashboard.read` | просмотр дашборда федерации |
| `registry.read` | просмотр единого реестра федерации |
| `judge_categories.review` | рассмотрение заявок на судейские категории |
| `rank_applications.review` | рассмотрение заявок на спортивные разряды |
| `deadlines.manage` | управление отчётными сроками |

Главный администратор из seed получает все эти permissions. Backend проверяет их по актуальным данным БД, поэтому назначение или удаление роли влияет на доступ независимо от `role_id` в старом JWT.

## Admin API

Все admin endpoint’ы требуют access-токен и соответствующий permission.

### Получение permissions

```http
GET /api/v1/admin/permissions
Authorization: Bearer <access_token>
```

Требуется `permissions.read`.

```json
{
  "data": [
    {
      "id": 1,
      "code": "roles.read",
      "description": "Просмотр ролей"
    }
  ]
}
```

### Получение ролей

```http
GET /api/v1/admin/roles
Authorization: Bearer <access_token>
```

Требуется `roles.read`.

```json
{
  "data": [
    {
      "id": 1,
      "code": "super_admin",
      "name": "Главный администратор",
      "is_system": true,
      "self_assignable": false,
      "permissions": [
        "permissions.read",
        "roles.assign",
        "roles.create",
        "roles.read"
      ]
    }
  ]
}
```

### Создание роли

```http
POST /api/v1/admin/roles
Authorization: Bearer <access_token>
```

Требуется `roles.create`.

```json
{
  "code": "coach",
  "name": "Тренер",
  "self_assignable": true,
  "permission_codes": [
    "roles.read"
  ]
}
```

Правила:

- `code` приводится к нижнему регистру;
- формат `code`: от 3 до 100 символов, первый символ — латинская буква, затем допустимы латинские буквы, цифры, `.`, `_`, `-`;
- `name` обязателен, максимум 100 символов;
- повторяющиеся permission-коды удаляются;
- все permission-коды должны существовать;
- `self_assignable: true` разрешает обычному пользователю выбрать роль через профиль.

Успешный ответ: `201 Created` с созданной ролью.

### Назначение роли пользователю

```http
POST /api/v1/admin/users/{userID}/roles
Authorization: Bearer <access_token>
```

Требуется `roles.assign`.

```json
{
  "role_id": 7
}
```

Успешный ответ: `204 No Content`. Повторное назначение той же роли безопасно и не создаёт дубликат. Admin API может назначить пользователю несколько ролей; итоговые permissions объединяются.

### Журнал аудита

```http
GET /api/v1/admin/audit-logs?limit=50&offset=0
Authorization: Bearer <access_token>
```

Требуется `roles.read`. Поддерживаются `limit` от 1 до 100 (по умолчанию 50) и неотрицательный `offset` (по умолчанию 0). События отсортированы от новых к старым.

```json
{
  "data": [
    {
      "id": 15,
      "event_type": "role.created",
      "actor_user_id": 1,
      "entity_type": "role",
      "entity_id": null,
      "method": "POST",
      "route": "/api/v1/admin/roles",
      "status": 201,
      "occurred_at": "2026-10-27T09:30:00Z"
    }
  ]
}
```

Сейчас записываются только успешные (`2xx`) операции: создание роли, назначение роли пользователю и изменение профиля. Ошибочные запросы и запросы без корректного access-токена не попадают в журнал.

Ошибки admin API:

- `400` — неверный JSON, user ID, role ID или permission-код;
- `401` — отсутствует или недействителен access-токен;
- `403` — нет требуемого permission;
- `404` — пользователь или роль не найдены;
- `409` — роль с таким `code` или `name` уже существует;
- `500` — внутренняя ошибка.

## Рекомендуемые TypeScript-типы

```ts
export interface TokenPair {
  access_token: string;
  refresh_token: string;
}

export interface ApiSuccess<T> {
  data: T;
}

export interface ApiError {
  message: string;
}

export interface RegisterFormFields {
  first_name: string;
  last_name: string;
  middle_name?: string;
  birth_date: string;
  email: string;
  password: string;
  personal_data_consent_file: File;
  parental_consent_document_file?: File;
}

export interface Role {
  id: number;
  code: string;
  name: string;
  is_system: boolean;
  self_assignable: boolean;
  permissions: string[];
}

export interface Permission {
  id: number;
  code: string;
  description: string;
}
```

## Рекомендуемый сценарий фронтенда

1. Собрать текстовые поля и документы регистрации в один `FormData`.
2. Рассчитать возраст на UI для показа поля родительского согласия, но окончательное решение оставить backend.
3. После регистрации или входа сохранить пару токенов.
4. Если в access-токене нет `role_id`, показать экран выбора роли.
5. При выборе ролей отправить непустой массив `role_ids` вместе с актуальным refresh-токеном.
6. Заменить токены значениями из ответа PATCH.
7. Для admin-раздела ориентироваться на ответы API и `403`, а не только на декодированный JWT.
8. При истечении access-токена один раз выполнить refresh и повторить исходный запрос.
9. Если refresh вернул `401`, очистить сессию и отправить пользователя на страницу входа.

## Текущие ограничения контракта

- Нет endpoint’а `GET /api/v1/profile` или `/me`.
- Нет публичного endpoint’а списка ролей с `self_assignable: true`.
- Нет отдельных публичных upload/delete endpoint’ов: документы загружаются только внутри регистрации.
- Пока нет API, которое отдаёт frontend ID прикреплённых документов; download endpoint готов для использования будущим profile/admin API.
- Refresh-токен возвращается в JSON, HttpOnly-cookie пока не используется.
- Нет CORS middleware для отдельного frontend-origin.
- Нет изменения/удаления ролей и снятия роли с пользователя.
- Пользовательский выбор ролей заменяет все назначенные роли; первый выбранный элемент становится активным в JWT, но отдельного endpoint для переключения активной роли без изменения набора пока нет.
- Нет списка пользователей для admin-интерфейса.
- Нет стабильного машинного `error_code`; фронт пока получает только текст `message`.
- OpenAPI/Swagger-схема ещё не подготовлена.
