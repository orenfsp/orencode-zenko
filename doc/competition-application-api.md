# API заявок на соревнования

Модуль управляет заявками спортсменов на конкретные зачёты `competition_event`. Он использует существующую таблицу `competition_application`; миграции и новые таблицы для модуля не требуются.

Все маршруты имеют префикс `/api/v1` и требуют заголовок `Authorization: Bearer <access_token>`. Успешный ответ имеет форму `{"data": ...}`, ошибка — `{"code":"...","message":"..."}`. Коды описаны в [общем контракте](./mvp-backend-api.md#формат-ошибок-основных-api). Неизвестные поля JSON отклоняются.

## Статусы

- `pending` — заявка ожидает решения;
- `approved` — заявка одобрена и учитывается модулем результатов;
- `rejected` — заявка отклонена;
- `withdrawn` — заявка отозвана спортсменом.

Для пары `(event_id, athlete_user_id)` разрешена только одна запись. Повторная подача, включая попытку создать новую запись после отзыва или отклонения, возвращает `409 Conflict`.

## Доступ и роли

Пользовательские маршруты используют `Authenticate`:

- спортсмен подаёт заявку за себя;
- тренер может указать `athlete_user_id` только при наличии подтверждённой и незавершённой связи `user_coach` по дисциплине зачёта;
- спортсмен видит собственные заявки, тренер — заявки подтверждённых подопечных;
- отозвать заявку может только сам спортсмен.

Административные маршруты дополнительно используют `Require("competition_applications.review")`. Разрешение входит в существующий RBAC seed и назначается ролям `admin` и `super_admin`. После обновления приложения для существующей БД нужно повторно выполнить:

```bash
cd backend
make seed-rbac
```

## Маршруты

| Метод и путь | Доступ | Назначение |
|---|---|---|
| `POST /competition-applications` | авторизованный пользователь | Подать заявку за себя или подтверждённого подопечного |
| `GET /competition-applications/my` | авторизованный пользователь | Свои заявки и доступные заявки подопечных |
| `GET /competition-applications/{id}` | спортсмен-владелец или подтверждённый тренер | Получить одну заявку |
| `POST /competition-applications/{id}/withdraw` | спортсмен-владелец | Отозвать ожидающую заявку до закрытия регистрации |
| `GET /admin/competition-applications` | `competition_applications.review` | Получить заявки для рассмотрения |
| `GET /admin/competition-applications/{id}` | `competition_applications.review` | Получить одну заявку для рассмотрения |
| `POST /admin/competition-applications/{id}/approve` | `competition_applications.review` | Одобрить ожидающую заявку |
| `POST /admin/competition-applications/{id}/reject` | `competition_applications.review` | Отклонить ожидающую заявку |

## Подача заявки

```http
POST /api/v1/competition-applications
Authorization: Bearer <access_token>
Content-Type: application/json

{"event_id":17}
```

`event_id` обязателен. Для подачи тренером добавляется ID спортсмена:

```json
{
  "event_id": 17,
  "athlete_user_id": 42
}
```

Backend проверяет:

1. существование зачёта и родительского соревнования;
2. статус соревнования `registration_open`;
3. попадание текущего времени в интервал `registration_starts_at`—`registration_ends_at`;
4. подтверждённую связь тренера со спортсменом по дисциплине, если заявка подаётся не за себя;
5. отсутствие заявки того же спортсмена на этот зачёт;
6. действующий антидопинговый сертификат существующим PostgreSQL-триггером.

Успешный ответ — `201 Created`. Заявка создаётся со статусом `pending`; `submitted_by_user_id` равен ID авторизованного пользователя.

## Формат заявки

```json
{
  "data": {
    "id": 101,
    "event_id": 17,
    "athlete_user_id": 42,
    "submitted_by_user_id": 42,
    "status": "pending",
    "submitted_at": "2026-10-28T08:00:00Z",
    "competition_id": 7,
    "competition_name": "Чемпионат Оренбургской области",
    "competition_start_at": "2026-11-10T06:00:00Z",
    "competition_end_at": "2026-11-10T15:00:00Z",
    "registration_starts_at": "2026-10-20T00:00:00Z",
    "registration_ends_at": "2026-11-05T20:00:00Z",
    "discipline_id": 2,
    "discipline_name": "Алгоритмическое программирование",
    "age_group_id": 3,
    "age_group_name": "Юниоры",
    "athlete_first_name": "Иван",
    "athlete_last_name": "Иванов",
    "athlete_email": "athlete@example.test",
    "submitter_first_name": "Иван",
    "submitter_last_name": "Иванов",
    "can_withdraw": true
  }
}
```

`reviewed_by_user_id` и `reviewed_at` появляются после решения. Поля возрастной группы и отчества могут отсутствовать. `can_withdraw` вычисляется относительно текущего пользователя, статуса и срока регистрации.

## Список пользователя и отзыв

```http
GET /api/v1/competition-applications/my
Authorization: Bearer <access_token>
```

Ответ содержит массив, отсортированный от новых заявок к старым. Для тренера в него входят заявки спортсменов, с которыми на момент запроса действует подтверждённая связь по дисциплине зачёта.

```http
POST /api/v1/competition-applications/101/withdraw
Authorization: Bearer <access_token>
```

Отзыв разрешён только владельцу заявки, только из статуса `pending` и не позднее `registration_ends_at`. Успешный ответ — `200 OK` с обновлённой заявкой.

## Рассмотрение

Список поддерживает параметры:

- `event_id` — положительный ID зачёта;
- `status` — `pending`, `approved`, `rejected` или `withdrawn`;
- `limit` — от 1 до 100, по умолчанию 50;
- `offset` — неотрицательное смещение.

```http
GET /api/v1/admin/competition-applications?event_id=17&status=pending&limit=50&offset=0
Authorization: Bearer <access_token>
```

Одобрение и отклонение не требуют JSON-тела:

```http
POST /api/v1/admin/competition-applications/101/approve
Authorization: Bearer <access_token>
```

```http
POST /api/v1/admin/competition-applications/101/reject
Authorization: Bearer <access_token>
```

Обе операции допустимы только для `pending`. Backend сохраняет новый статус, `reviewed_by_user_id` и `reviewed_at`. Повторное решение возвращает `409 Conflict`. Существующий триггер создаёт уведомление спортсмену при изменении статуса.

После одобрения заявка автоматически попадает в `ListApprovedResultApplications`, поэтому её можно включить в таблицу результатов зачёта. Дашборд читает только `pending`, следовательно рассмотренная или отозванная заявка исчезает из очереди без дополнительной синхронизации.

## Коды ответа

- `200 OK` — успешное чтение или изменение;
- `201 Created` — заявка создана;
- `400 Bad Request` — неверный ID, фильтр или JSON;
- `401 Unauthorized` — access-токен отсутствует или недействителен;
- `403 Forbidden` — нет доступа к спортсмену/заявке или административного permission;
- `404 Not Found` — зачёт либо заявка не найдены;
- `409 Conflict` — дубликат, повторное решение или недопустимое изменение статуса;
- `422 Unprocessable Entity` — регистрация закрыта либо нет действующего антидопингового сертификата;
- `500 Internal Server Error` — непредвиденная ошибка backend/PostgreSQL.

## Ручная проверка

```bash
curl -X POST http://localhost:8080/api/v1/competition-applications \
  -H "Authorization: Bearer $ATHLETE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"event_id":17}'

curl http://localhost:8080/api/v1/competition-applications/my \
  -H "Authorization: Bearer $ATHLETE_TOKEN"

curl "http://localhost:8080/api/v1/admin/competition-applications?status=pending" \
  -H "Authorization: Bearer $ADMIN_TOKEN"

curl -X POST http://localhost:8080/api/v1/admin/competition-applications/101/approve \
  -H "Authorization: Bearer $ADMIN_TOKEN"
```

