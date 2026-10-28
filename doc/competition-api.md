# Competition API

Все маршруты используют префикс `/api/v1` и формат ответа проекта:

```json
{
  "data": {}
}
```

Ошибка возвращается как `{"code":"...","message":"..."}`; коды описаны в [общем контракте](./mvp-backend-api.md#формат-ошибок-основных-api). Неизвестные поля JSON и второе значение после основного JSON-объекта отклоняются с `400 Bad Request`.

## Административный жизненный цикл

Список собственных соревнований администратора:

```http
GET /api/v1/admin/competitions?status=draft
Authorization: Bearer <access_token>
```

Доступны статусы `draft`, `under_review`, `published`, `cancelled`. Получить одно своё соревнование можно через `GET /api/v1/admin/competitions/{id}`.

Переходы выполняются отдельными POST-ручками:

```http
POST /api/v1/admin/competitions/{id}/submit
POST /api/v1/admin/competitions/{id}/publish
POST /api/v1/admin/competitions/{id}/cancel
POST /api/v1/admin/competitions/{id}/return
```

Переходы: `draft -> under_review`, `under_review -> published`, `under_review|published -> cancelled`, `under_review -> draft` (возврат на доработку). История доступна через `GET /api/v1/admin/competitions/{id}/history`.


## Маршруты и доступ

| Метод и путь | Доступ | Назначение |
|---|---|---|
| `GET /calendar/competitions?from=2026-11-01&to=2026-12-01` | публичный | Компактный календарь соревнований за период |
| `GET /competitions?limit=50&offset=0` | публичный | Список публичных соревнований |
| `GET /competitions/{id}` | публичный | Публичное соревнование по ID |
| `POST /competitions` | access-токен и `competitions.create` | Создать черновик |
| `PATCH /competitions/{id}` | access-токен и `competitions.update` | Частично изменить черновик |

Публичные GET показывают соревнования в статусах `published`, `registration_open`, `registration_closed`, `ongoing`, `results_pending` и `completed`. Статусы `draft` и `under_review` через публичные маршруты недоступны. Поэтому каждое соревнование из календаря можно получить через `GET /competitions/{id}`.

Для защищённых запросов нужен заголовок `Authorization: Bearer <access_token>`. Значение `created_by` backend всегда берёт из проверенного JWT-контекста; клиент не может его задать.

## Публичный календарь

```http
GET /api/v1/calendar/competitions?from=2026-11-01&to=2026-12-01&discipline_id=2&level_id=4&limit=50&offset=0
```

Авторизация не требуется. Параметры `from` и `to` обязательны и задают полуоткрытый интервал `[from, to)`: начало включается, конец не включается. Например, `from=2026-11-01&to=2026-12-01` возвращает соревнования, пересекающие ноябрь, но не декабрь.

Границы можно передать в одном из форматов:

- `YYYY-MM-DD` — полночь указанной даты в часовом поясе `Asia/Yekaterinburg`;
- RFC3339 с обязательным часовым поясом, например `2026-11-01T00:00:00+05:00` или `2026-10-31T19:00:00Z`.

В URL символ `+` в смещении нужно кодировать как `%2B`. Период должен быть непустым: `from` обязан быть раньше `to`.

Опциональные параметры:

- `discipline_id` — положительный ID дисциплины;
- `level_id` — положительный ID уровня соревнования;
- `limit` — от 1 до 100, по умолчанию 50;
- `offset` — неотрицательное число, по умолчанию 0.

Соревнование попадает в календарь, если его интервал пересекает запрошенный период: `start_at < to` и `end_at > from`. Результат отсортирован по `start_at`, затем по `id`.

```json
{
  "data": [
    {
      "id": 7,
      "name": "Кубок Оренбуржья",
      "discipline": {"id": 2, "name": "Алгоритмическое программирование"},
      "competition_type": {"id": 3, "name": "Кубок"},
      "competition_level": {"id": 4, "code": "regional", "name": "Региональный"},
      "format": "mixed",
      "start_at": "2026-11-10T08:00:00Z",
      "end_at": "2026-11-11T16:00:00Z",
      "venue_address": "Оренбург, ул. Советская, 1",
      "online_url": "https://contest.example.org",
      "status": "published"
    }
  ]
}
```

## Создание соревнования

```http
POST /api/v1/competitions
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "name": "Кубок Оренбуржья",
  "description": "Региональное соревнование",
  "discipline_id": 2,
  "competition_type_id": 3,
  "competition_level_id": 4,
  "format": "mixed",
  "start_at": "2026-11-10T08:00:00Z",
  "end_at": "2026-11-11T16:00:00Z",
  "registration_starts_at": "2026-10-01T00:00:00Z",
  "registration_ends_at": "2026-11-01T00:00:00Z",
  "venue_address": "Оренбург, ул. Советская, 1",
  "online_url": "https://contest.example.org",
  "calendar_plan_id": null,
  "calendar_plan_number": null,
  "regulation_url": "https://example.org/regulation.pdf"
}
```

Успешный ответ имеет статус `201 Created`. Новая запись всегда получает статус `draft`; прислать `status` или `created_by` в запросе нельзя.

Формат может быть `offline`, `online` или `mixed`. Для `offline` и `mixed` обязателен адрес, для `online` и `mixed` — корректный HTTP(S)-URL. Начало должно предшествовать окончанию, период регистрации должен быть корректным и завершаться не позже начала соревнования. ID дисциплины, типа, уровня и календарного плана проверяются по справочникам БД.

## Частичное изменение

```http
PATCH /api/v1/competitions/7
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "name": "Новое название",
  "description": null,
  "end_at": "2026-11-12T16:00:00Z"
}
```

PATCH поддерживает настоящую частичную семантику:

- отсутствующее поле сохраняет прежнее значение;
- явный `null` очищает nullable-поле;
- `null` для обязательного поля возвращает `400`;
- пустой объект и неизвестное поле возвращают `400`;
- после объединения старых и новых значений backend проверяет весь итоговый объект;
- изменить `id`, `created_by`, `created_at`, `updated_at` или `status` нельзя;
- редактируется только запись со статусом `draft`; статус проверяется также атомарно в SQL при записи;
- успешное обновление устанавливает `updated_at = now()`, но оставляет статус `draft`.

Изменяемые поля: `name`, `description`, `discipline_id`, `competition_type_id`, `competition_level_id`, `format`, `start_at`, `end_at`, `registration_starts_at`, `registration_ends_at`, `venue_address`, `online_url`, `calendar_plan_id`, `calendar_plan_number`, `regulation_url`.

Если очищается `calendar_plan_id`, связанный `calendar_plan_number` также нужно очистить в том же запросе:

```json
{
  "calendar_plan_id": null,
  "calendar_plan_number": null
}
```

Успешный ответ имеет статус `200 OK` и содержит полный объект соревнования с раскрытыми справочниками:

```json
{
  "data": {
    "id": 7,
    "name": "Новое название",
    "description": null,
    "discipline": {"id": 2, "name": "Алгоритмическое программирование"},
    "competition_type": {"id": 3, "name": "Кубок"},
    "competition_level": {"id": 4, "code": "regional", "name": "Региональный"},
    "format": "mixed",
    "start_at": "2026-11-10T08:00:00Z",
    "end_at": "2026-11-12T16:00:00Z",
    "registration_starts_at": "2026-10-01T00:00:00Z",
    "registration_ends_at": "2026-11-01T00:00:00Z",
    "venue_address": "Оренбург, ул. Советская, 1",
    "online_url": "https://contest.example.org",
    "status": "draft",
    "calendar_plan": null,
    "regulation_url": "https://example.org/regulation.pdf"
  }
}
```

## Ошибки

- `400 Bad Request` — неверный ID/JSON, пустой PATCH, неизвестное или запрещённое поле, нарушение правил полей, дат, URL или справочников;
- `401 Unauthorized` — access-токен отсутствует или недействителен;
- `403 Forbidden` — токен действителен, но у пользователя нет необходимого permission;
- `404 Not Found` — соревнование не существует; публичный GET также отвечает 404 для черновика;
- `409 Conflict` — имя уже занято либо PATCH направлен на соревнование не в статусе `draft`;
- `500 Internal Server Error` — непредвиденная ошибка backend или PostgreSQL.

## Настройка RBAC

Маршруты используют существующие `Authenticate`, `Require` и таблицы `permission`, `role_permission`, `role_user`. Нужны два permission-кода:

- `competitions.create` — создание соревнования;
- `competitions.update` — изменение черновика.

Они намеренно не добавлены миграцией и не внесены в модуль авторизации. Минимальная отдельная задача для владельца RBAC:

1. Добавить оба определения в `rbac.DefaultPermissions`.
2. Запустить существующий `cmd/seed-rbac`: он вызовет `UpsertPermission` и назначит новые permissions системной роли `super_admin`.
3. Для роли сотрудника федерации получить ID через `GetPermissionIDByCode` и назначить оба permission существующим `AssignRolePermission`. Если роль создаётся заново, её можно создать через Admin API с этими кодами в `permission_codes`.

Миграция и изменение JWT/middleware для этого не нужны. После назначения текущий `HasPermission` начнёт разрешать маршруты по актуальным данным БД. Пока permissions отсутствуют или не назначены, авторизованный пользователь корректно получает `403`.

## Примеры curl

```bash
curl "http://localhost:8080/api/v1/calendar/competitions?from=2026-11-01&to=2026-12-01&discipline_id=2"

curl http://localhost:8080/api/v1/competitions

curl http://localhost:8080/api/v1/competitions/7

curl -X POST http://localhost:8080/api/v1/competitions \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  --data @competition.json

curl -X PATCH http://localhost:8080/api/v1/competitions/7 \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Новое название","description":null}'
```
