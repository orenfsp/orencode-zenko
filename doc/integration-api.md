# API внешних интеграций

Документ описывает серверное подключение соревновательных платформ, сайтов организаций и сервисов партнёров. Отдельный frontend для интеграции не требуется.

## 1. Базовые правила

- Базовый URL примеров: `http://localhost:8080`.
- Все тела запросов и ответов используют UTF-8 JSON.
- Секреты и access-токены должны передаваться только через HTTPS.
- Интеграционный JWT имеет `principal_type: integration` и не является токеном пользователя.
- Refresh token не выдаётся. После истечения `expires_in` клиент повторно получает access token.
- Отключение интеграционного клиента или его владельца немедленно запрещает доступ, даже если JWT ещё не истёк.
- Интеграционный клиент видит только разрешённые ему операции.

Доступные permissions:

| Permission | Назначение |
|---|---|
| `integration.results.import` | получение таблицы, импорт, расчёт, утверждение результатов и генерация протокола |
| `integration.webhooks.manage` | создание, просмотр и удаление собственных webhook-подписок |

## 2. Создание интеграционного клиента

Клиент создаётся CLI-командой без административного frontend. Предварительно примените миграции и убедитесь, что пользователь-владелец существует и активен.

```bash
docker compose run --rm \
  --entrypoint /seed-integration \
  -e INTEGRATION_CLIENT_ID=partner-system \
  -e INTEGRATION_CLIENT_SECRET='replace-with-strong-secret' \
  -e INTEGRATION_CLIENT_NAME='Партнёрская система' \
  -e INTEGRATION_OWNER_EMAIL='admin@example.com' \
  -e INTEGRATION_PERMISSIONS='integration.results.import,integration.webhooks.manage' \
  server
```

Повторный запуск с тем же `INTEGRATION_CLIENT_ID` обновляет название, владельца, секрет и полностью заменяет permissions. Секрет должен содержать от 16 до 72 байт; в БД сохраняется только bcrypt-хеш.

Владелец используется для атрибуции утверждения результатов в существующей модели протоколов. Доступ интеграции при этом определяется её собственными permissions, а не правами владельца.

## 3. Получение access token

```http
POST /api/v1/integrations/token
Content-Type: application/json
```

```json
{
  "client_id": "partner-system",
  "client_secret": "replace-with-strong-secret"
}
```

Успешный ответ:

```json
{
  "data": {
    "access_token": "eyJ...",
    "token_type": "Bearer",
    "expires_in": 900,
    "permissions": [
      "integration.results.import",
      "integration.webhooks.manage"
    ]
  }
}
```

Во всех защищённых запросах:

```http
Authorization: Bearer <access_token>
```

Неверные реквизиты возвращают `401 invalid_client`. Интеграционный токен нельзя использовать вместо пользовательского токена на `/api/v1/admin/*`.

## 4. Webhooks

Поддерживаемые события:

| Event | Когда создаётся |
|---|---|
| `competition.published` | соревнование опубликовано |
| `application.approved` | заявка на соревнование одобрена |
| `protocol.published` | результаты утверждены либо PDF-протокол загружен или сформирован импортом |

### 4.1. Создание подписки

```http
POST /api/v1/integrations/webhooks
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "url": "https://partner.example/api/zenko/webhook",
  "signing_secret": "webhook-signing-secret-123",
  "events": [
    "competition.published",
    "application.approved",
    "protocol.published"
  ]
}
```

В production разрешён только HTTPS. HTTP допускается для `localhost` и loopback-адресов. Секрет подписи должен содержать от 16 до 200 символов.

Ответ `201 Created` не возвращает `signing_secret`:

```json
{
  "data": {
    "id": 4,
    "url": "https://partner.example/api/zenko/webhook",
    "events": ["application.approved", "competition.published", "protocol.published"],
    "is_active": true,
    "created_at": "2026-10-28T12:00:00Z"
  }
}
```

### 4.2. Получение и удаление подписок

```http
GET /api/v1/integrations/webhooks
DELETE /api/v1/integrations/webhooks/{subscriptionID}
Authorization: Bearer <access_token>
```

Удаление возвращает `204 No Content`. Клиент может читать и удалять только собственные подписки.

### 4.3. Формат доставки

```json
{
  "type": "protocol.published",
  "occurred_at": "2026-10-28T12:10:15.123Z",
  "data": {
    "competition_id": 10,
    "event_id": 20
  }
}
```

Заголовки:

```http
Content-Type: application/json
X-Zenko-Event-ID: 153
X-Zenko-Event-Type: protocol.published
X-Zenko-Signature: sha256=<hex_hmac>
```

`X-Zenko-Signature` — HMAC-SHA256 от точных байтов тела запроса с ключом `signing_secret`. Сравнивайте подпись constant-time функцией и используйте `X-Zenko-Event-ID` для идемпотентности.

Получатель должен вернуть любой `2xx`. При сетевой ошибке или другом статусе backend выполняет до пяти попыток с задержками приблизительно 5, 10, 20, 40 и 80 секунд. После пятой ошибки доставка получает статус `failed`.

Пример проверки подписи на Go:

```go
func validSignature(body []byte, header, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(body)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(expected), []byte(header))
}
```

## 5. Импорт результатов

Импорт использует существующий result-модуль. Таблица зачёта и одобренные заявки должны быть подготовлены заранее. Успешный импорт проходит прикладной сценарий:

1. сохранение строк;
2. проверка и расчёт;
3. утверждение результатов;
4. формирование PDF-протокола;
5. создание события `protocol.published`.

### 5.1. Получение таблицы и `entry_id`

```http
GET /api/v1/integrations/competitions/{competitionID}/events/{eventID}/results
Authorization: Bearer <access_token>
```

Ответ совпадает с административной таблицей результатов и содержит:

- `scoring_mode` — `algorithmic` или `manual`;
- `revision` — версия для ручного импорта;
- `entries[].id` — значение `entry_id` для импорта;
- `entries[].participants` — заявки и пользователи, связанные со строкой.

### 5.2. Алгоритмический зачёт

```http
POST /api/v1/integrations/competitions/10/events/20/results/import
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "entries": [
    {
      "entry_id": 101,
      "result_status": "finished",
      "tasks": [
        {
          "task_number": 1,
          "status": "solved",
          "solve_time_deciseconds": 1250,
          "negative_attempts": 1
        },
        {
          "task_number": 2,
          "status": "not_attempted",
          "solve_time_deciseconds": null,
          "negative_attempts": 0
        }
      ]
    }
  ]
}
```

### 5.3. Ручной зачёт

```json
{
  "revision": 4,
  "entries": [
    {
      "entry_id": 101,
      "result_status": "finished",
      "place": 1,
      "custom_values": {
        "points": 98.5,
        "comment": "Подтверждено внешней системой"
      }
    }
  ]
}
```

Поля `custom_values` должны соответствовать активному шаблону зачёта. `revision` берётся из GET и защищает от перезаписи параллельных изменений.

Успешный ответ:

```json
{
  "data": {
    "status": "approved",
    "approved_at": "2026-10-28T12:20:00Z",
    "protocol_url": "/api/v1/competitions/10/events/20/protocol"
  }
}
```

Повторный запрос после успешного утверждения возвращает существующий протокол и не создаёт его заново.

## 6. Ошибки

```json
{
  "code": "forbidden",
  "message": "integration permission denied"
}
```

Основные статусы:

- `400` — неверный JSON, ID или структура результатов;
- `401` — отсутствует, истёк или неверен интеграционный токен;
- `403` — клиенту не выдан требуемый permission;
- `404` — зачёт, таблица, строка или webhook-подписка не найдены;
- `409` — результаты уже изменены, состав заявок изменился или URL подписки уже занят;
- `422` — режим или дисциплина не поддерживают переданный формат;
- `500` — внутренняя ошибка.

## 7. Минимальный пример подключения

```bash
TOKEN=$(curl -s http://localhost:8080/api/v1/integrations/token \
  -H 'Content-Type: application/json' \
  -d '{"client_id":"partner-system","client_secret":"replace-with-strong-secret"}' \
  | jq -r '.data.access_token')

curl http://localhost:8080/api/v1/integrations/competitions/10/events/20/results \
  -H "Authorization: Bearer $TOKEN"
```

Для production используйте HTTPS, храните `client_secret` и `signing_secret` в менеджере секретов и задавайте таймауты HTTP-клиента.
