# Competition Event API

`competition_event` — зачёт внутри одного соревнования. Даты, место, формат, статус и организатор принадлежат родительскому Competition и в Event не дублируются. В текущей MVP-модели дисциплина Event всегда берётся backend из Competition.

Все маршруты используют префикс `/api/v1`. Успешный JSON имеет вид `{"data": ...}`, ошибка — `{"message":"..."}`. Неизвестные поля JSON отклоняются.

## Маршруты и permissions

| Метод и путь | Доступ | Назначение |
|---|---|---|
| `GET /competitions/{competitionID}/events` | публичный | Список зачётов публичного соревнования |
| `GET /competitions/{competitionID}/events/{eventID}` | публичный | Один зачёт публичного соревнования |
| `GET /admin/competitions/{competitionID}/events` | `competition_events.read` | Список зачётов, включая черновик |
| `GET /admin/competitions/{competitionID}/events/{eventID}` | `competition_events.read` | Один зачёт, включая черновик |
| `POST /competitions/{competitionID}/events` | `competition_events.create` | Создать зачёт в черновике |
| `PATCH /competitions/{competitionID}/events/{eventID}` | `competition_events.update` | Изменить возрастную группу зачёта |
| `DELETE /competitions/{competitionID}/events/{eventID}` | `competition_events.delete` | Удалить зачёт без заявок |

Защищённым маршрутам нужен `Authorization: Bearer <access_token>`. Они используют общие middleware `Authenticate` и `Require`; permission-коды регистрируются существующим `cmd/seed-rbac` и назначаются системной роли `super_admin`.

Публичные GET возвращают Events только для Competition в статусах `published`, `registration_open`, `registration_closed`, `ongoing`, `results_pending`, `completed` — тот же набор использует публичный календарь. `draft` и `under_review` скрыты и дают `404 Not Found`.

Каждая операция с одним Event проверяет одновременно `competitionID` и `eventID`. Поэтому ID зачёта другого соревнования также даёт 404.

## Формат Event

```json
{
  "data": {
    "id": 9,
    "competition_id": 7,
    "discipline": {
      "id": 2,
      "name": "Алгоритмическое программирование"
    },
    "age_group": {
      "id": 3,
      "name": "Юниоры"
    },
    "protocol_url": null
  }
}
```

`age_group` может быть `null`. `protocol_url` возвращается в GET, но POST/PATCH его не принимают: загрузка и публикация протоколов относятся к будущему модулю результатов.

## Создание

```http
POST /api/v1/competitions/7/events
Authorization: Bearer <access_token>
Content-Type: application/json

{"age_group_id":3}
```

`age_group_id` необязателен; если задан, справочник `age_group` должен содержать запись. `competition_id` берётся только из URL, а `discipline_id` — из заблокированной строки родительского Competition. Попытка передать эти поля в JSON отклоняется как неизвестное поле. Успешный ответ: `201 Created` и полный Event.

Создавать Event можно только у Competition со статусом `draft`. Repository начинает транзакцию, блокирует строку Competition через `SELECT ... FOR UPDATE`, проверяет статус и только затем выполняет INSERT. Будущая операция публикации должна блокировать ту же строку Competition; тогда публикация и изменение Events не смогут пройти одновременно на противоречивых состояниях.

## Частичное изменение

```http
PATCH /api/v1/competitions/7/events/9
Authorization: Bearer <access_token>
Content-Type: application/json

{"age_group_id":4}
```

Единственное изменяемое поле — `age_group_id`. Чтобы очистить его, передайте явный `null`:

```json
{"age_group_id":null}
```

Отсутствующее поле не означает очистку: пустой PATCH возвращает 400. `id`, `competition_id`, `discipline_id` и `protocol_url` менять нельзя. Если новая возрастная группа отличается от текущей и на Event уже существуют заявки, ответ — `409 Conflict`. Повторная установка того же значения считается безопасной и не конфликтует с заявками.

Repository в одной транзакции блокирует сначала Competition, затем Event. Это обеспечивает стабильный порядок блокировок и не оставляет гонки между проверками статуса/заявок и UPDATE.

## Удаление

```http
DELETE /api/v1/competitions/7/events/9
Authorization: Bearer <access_token>
```

Удаление разрешено только в `draft` и только если у Event нет заявок. Заявки и результаты автоматически не удаляются. Успешный ответ: `204 No Content` без JSON-тела. При наличии заявок — `409 Conflict`.

## Ошибки

- `400 Bad Request` — неверные ID/JSON, пустой PATCH, неизвестное поле, неположительный или отсутствующий в справочнике `age_group_id`;
- `401 Unauthorized` — access-токен отсутствует или недействителен;
- `403 Forbidden` — нет требуемого permission;
- `404 Not Found` — Competition/Event не найден, Event принадлежит другому Competition либо публичный Competition скрыт;
- `409 Conflict` — Competition не в `draft` или Event связан с заявками;
- `500 Internal Server Error` — непредвиденная ошибка backend/PostgreSQL без раскрытия внутреннего текста.

## Ручная проверка

```bash
curl http://localhost:8080/api/v1/competitions/7/events

curl http://localhost:8080/api/v1/competitions/7/events/9

curl http://localhost:8080/api/v1/admin/competitions/7/events \
  -H "Authorization: Bearer $ACCESS_TOKEN"

curl -X POST http://localhost:8080/api/v1/competitions/7/events \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"age_group_id":3}'

curl -X PATCH http://localhost:8080/api/v1/competitions/7/events/9 \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"age_group_id":null}'

curl -i -X DELETE http://localhost:8080/api/v1/competitions/7/events/9 \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

Публикация Competition остаётся отдельной операцией. Перед переводом из `draft` она должна проверять готовность Events и использовать блокировку той же строки Competition, что и Event Repository.
