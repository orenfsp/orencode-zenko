# Backend API пяти направлений MVP

Документ описывает только фактически реализованные backend-возможности. Он дополняется одновременно с каждым этапом MVP.

## Формат ошибок основных API

Ошибки авторизации и RBAC, соревнований и зачётов, заявок на соревнования, результатов и заявок на разряд возвращаются как JSON:

```json
{"code":"registration_closed","message":"регистрация на соревнование закрыта"}
```

`code` — стабильный машинный идентификатор, `message` — текст для пользователя, сохранённый для совместимости. Опциональный `details` содержит структурированные детали, когда они предусмотрены; пустое поле не добавляется. HTTP-статусы и успешные ответы не менялись.

| Код | HTTP | Значение |
|---|---|---|
| `invalid_request` | 400 | Некорректный запрос или данные |
| `unauthorized` | 401 | Нет действительной авторизации |
| `forbidden` | 403 | Недостаточно прав |
| `not_found` | 404 | Объект не найден |
| `conflict` | 409 | Конфликт операции |
| `invalid_competition_status` | 409 | Операция недоступна в текущем статусе соревнования |
| `application_already_exists` | 409 | Заявка уже существует |
| `application_already_reviewed` | 409 | Заявка рассмотрена или недоступна для изменения |
| `result_not_approved` | 409 | Зачёт не утверждён |
| `payload_too_large` | 413 | Файл или запрос слишком велик |
| `unsupported_media_type` | 415 | Формат файла не поддерживается |
| `validation_error` | 422 | Данные не проходят бизнес-проверку |
| `registration_closed` | 422 | Регистрация закрыта |
| `antidoping_required` | 422 | Нужен действующий антидопинговый сертификат |
| `internal_error` | 500 | Внутренняя ошибка без раскрытия деталей |

Например, отсутствие токена возвращает `401 {"code":"unauthorized","message":"требуется access-токен"}`, а повторная заявка на соревнование — `409` с `application_already_exists`. Заявки на разряд с чужим или неподтверждённым основанием пока возвращают общий `invalid_request`: сервис не различает эти причины отдельными ошибками.

Ограничение MVP: остальные обработчики, включая `user_coach`, `rank`, `judgecategory`, `file`, `antidoping`, `calendarplan`, `dashboard`, `registry`, `rating`, `team`, `deadline`, `municipality`, `reference`, `evsk`, `audit`, `notification` и модуль справок, могут возвращать прежний формат или текст. Ошибки middleware RBAC на их защищённых маршрутах уже имеют новый формат. Ответы стандартного роутера на неизвестные маршруты не унифицированы.

## Дашборд федерации

### `GET /api/v1/admin/dashboard`

Требуется access-токен и разрешение `dashboard.read`. Разрешение добавлено в стандартный RBAC seed и автоматически назначается роли `super_admin` при повторном запуске `cmd/seed-rbac`.

Ответ `200 OK`:

```json
{
  "data": {
    "metrics": {
      "athletes": 14,
      "competitions": 3,
      "participants": 9,
      "active_ranks": 7
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

Правила расчёта:

- `athletes` — уникальные пользователи с ролью `participant`;
- `competitions` — все соревнования, включая черновики;
- `participants` — уникальные спортсмены с одобренной заявкой на соревнование;
- `active_ranks` — присвоения без срока окончания или со сроком не раньше текущей даты;
- списки ограничены десятью ближайшими записями;
- истекающими считаются разряды и судейские категории, срок которых заканчивается в ближайшие 60 дней;
- `deadlines` на этом этапе содержит сроки окончания регистрации опубликованных соревнований и соревнований с открытой регистрацией.

Пустые наборы всегда сериализуются как `[]`. Возможные ошибки: `401` без корректного access-токена, `403` без `dashboard.read`, `500` при ошибке БД.

Заявки на судейскую категорию появятся после добавления соответствующего сценария и таблицы на этапе «Судьи». Отчётные сроки будут включены на этапе «Контроль сроков». До этих этапов соответствующие массивы честно остаются пустыми, фиктивные данные не используются.

## Поиск спортсменов тренером

`GET /api/v1/coach/participants` требует `Authorization: Bearer <access_token>` и назначенную пользователю роль `coach`. Административное разрешение `registry.read` не требуется. Поиск охватывает всех активных пользователей с ролью `participant`, включая спортсменов без связи с тренером; сам тренер из результата исключается.

Параметры: `search` — часть имени, фамилии, отчества или ФИО без учёта регистра; `discipline_id` — положительный ID дисциплины из `user_discipline`; `limit` — от 1 до 100, по умолчанию 20; `offset` — неотрицательное смещение, по умолчанию 0. Пустой поиск возвращает всех подходящих спортсменов. Сортировка стабильна по фамилии, имени и ID; `total` — число результатов до пагинации, в том числе при пустой странице.

```json
{
  "data": {
    "items": [{
      "id": 42,
      "first_name": "Иван",
      "last_name": "Иванов",
      "middle_name": null,
      "organization": "Спортивная школа",
      "municipality": "Оренбург",
      "disciplines": [{"id": 1, "name": "Программирование алгоритмическое"}],
      "coach_relations": [{"discipline_id": 1, "status": "confirmed"}]
    }],
    "total": 1,
    "limit": 20,
    "offset": 0
  }
}
```

`coach_relations` содержит последнюю связь с текущим тренером по каждой дисциплине, причём действующая `pending` или `confirmed` имеет приоритет над завершённой. Массивы без записей возвращаются как `[]`. Email, дата рождения и документы не возвращаются. Для найденного `id` тренер использует существующий `POST /api/v1/coach/user-coaches` с телом `{"participant_user_id":42,"discipline_id":1}`; спортсмен подтверждает связь через `PATCH /api/v1/user-coaches/{id}/confirm`.

```bash
curl -G http://localhost:8080/api/v1/coach/participants \
  -H "Authorization: Bearer $COACH_TOKEN" \
  --data-urlencode 'search=Иванов' \
  --data-urlencode 'discipline_id=1' \
  --data-urlencode 'limit=20' \
  --data-urlencode 'offset=0'
```

Ошибки: `400` — неверные фильтры или пагинация, `401` — отсутствует или недействителен токен, `403` — у пользователя нет роли тренера, `500` — внутренняя ошибка. Новых разрешений и миграций не требуется.

## Единый реестр

Оба endpoint требуют access-токен и разрешение `registry.read`. Разрешение входит в стандартный RBAC seed.

### `GET /api/v1/admin/registry/users`

Возвращает спортсменов, тренеров, судей и других пользователей с доступными сведениями. Параметры:

- `search` — часть ФИО без учёта регистра;
- `discipline_id`, `organization_id`, `rank_id`, `age_group_id`, `municipality_id` — положительные ID;
- `role` — код роли, например `participant`, `coach` или `judge`;
- `limit` — от 1 до 100, по умолчанию 50;
- `offset` — неотрицательное смещение.

Фильтры можно совмещать. Разряды в ответе и фильтр `rank_id` учитывают только действующие присвоения. Возраст вычисляется на текущую дату и сопоставляется с границами выбранной записи `age_group`.

```json
{
  "data": {
    "items": [
      {
        "id": 17,
        "last_name": "Иванов",
        "first_name": "Иван",
        "middle_name": null,
        "birth": "2004-05-20T00:00:00Z",
        "organization": {"id": 2, "name": "Спортивная школа"},
        "municipality": {"id": 1, "name": "Оренбург"},
        "roles": [{"id": 3, "code": "judge", "name": "Судья"}],
        "disciplines": [{"id": 1, "name": "Программирование продуктовое"}],
        "ranks": [],
        "judge_category": null
      }
    ],
    "total": 1,
    "limit": 50,
    "offset": 0
  }
}
```

### `GET /api/v1/admin/registry/organizations`

Параметры: `search`, `municipality_id`, `limit`, `offset`. Ответ использует такую же пагинацию и возвращает `id`, `name`, `organization_type`, `municipality`.

Стабильная сортировка: пользователи по ФИО и ID, организации по названию и ID. Пустой результат содержит `items: []` и `total: 0`. Ошибки: `400` для некорректных фильтров, `401`, `403`, `500`.

## Судьи и судейские категории

### Миграция

`015_judge_category_applications.up.sql` добавляет таблицу `judge_category_application`. Она хранит только заявки и решения. Подтверждённые категории по-прежнему хранятся в `user_judge_category`. Для одного судьи, категории и типа заявки одновременно разрешена только одна заявка со статусом `pending`.

### Заявки судьи

- `POST /api/v1/judge-category-applications` — подать заявку; требуется access-токен и роль `judge`.
- `GET /api/v1/judge-category-applications/mine` — история собственных заявок.
- `GET /api/v1/judge-category/current` — актуальная подтверждённая категория или `data: null`.

Тело создания:

```json
{
  "category_id": 2,
  "application_type": "assignment"
}
```

`application_type`: `assignment` для первичного присвоения, `confirmation` для подтверждения. Повторная активная заявка возвращает `409`.

### Рассмотрение федерацией

- `GET /api/v1/admin/judge-category-applications?limit=50&offset=0` — ожидающие заявки;
- `PATCH /api/v1/admin/judge-category-applications/{id}` — одобрение или отклонение.

Оба endpoint требуют `judge_categories.review`.

Одобрение:

```json
{
  "status": "approved",
  "assigned_at": "2026-10-27",
  "expires_at": "2028-10-27",
  "order_number": "42-к"
}
```

Отклонение:

```json
{
  "status": "rejected",
  "rejection_reason": "Не хватает подтверждающих документов"
}
```

Одобрение и создание записи `user_judge_category` выполняются в одной транзакции. `assigned_at` обязателен для одобрения; `expires_at` не может быть раньше даты присвоения; причина обязательна при отклонении.

### Судейская коллегия

Существующие назначение и снятие судьи не изменены и по-прежнему доступны только для черновика соревнования:

- `POST /api/v1/competitions/{id}/judges`;
- `DELETE /api/v1/competitions/{id}/judges/{assignmentId}`.

Добавлены:

- `GET /api/v1/admin/competitions/{id}/judges` с `competition_judges.manage` — полный состав, включая черновик;
- `GET /api/v1/judge/competitions` — собственные назначения судьи («Мои соревнования»), требуется access-токен.

Публичный `GET /api/v1/competitions/{id}/judges` сохранён без изменений и не раскрывает состав непубличного соревнования.

## Спортивные разряды

### Миграция и связь с результатом

`016_rank_application_basis.up.sql` дополняет существующую `rank_application` полями:

- `application_type` (`assignment` или `confirmation`);
- `competition_id` и `result_id` — основание заявки;
- `assigned_at`, `expires_at`, `order_number` — реквизиты решения.

Новая система разрядов не создаётся. При одобрении существующий триггер записывает результат в `user_rank_discipline`. Ручное рассмотрение и автоматическое присвоение через `competition_result.awarded_rank_id` используют один `result_id` и сходятся в одной записи благодаря существующему уникальному индексу.

### Подача заявки

`POST /api/v1/rank-applications` требует access-токен. Спортсмен может подать заявку за себя; подтверждённый тренер — за своего подопечного по той же дисциплине.

Для выбора основания спортсмен получает историю через `GET /api/v1/profile/competition-results` с access-токеном. Необязательный `discipline_id` ограничивает результаты выбранной дисциплиной на стороне БД; он должен быть положительным целым числом. Без фильтра возвращается вся собственная история, включая результаты, которые ещё не являются допустимым основанием.

```json
{
  "data": [{
    "result_id": 25,
    "competition_id": 8,
    "competition_name": "Первенство области",
    "event_id": 12,
    "discipline_id": 1,
    "discipline_name": "Программирование алгоритмическое",
    "start_at": "2026-10-27T09:00:00Z",
    "place": 2,
    "score": 5,
    "result_status": "finished",
    "protocol_url": "/api/v1/competitions/8/events/12/protocol"
  }]
}
```

`result_id` — ID записи `competition_result` конкретного спортсмена, в том числе для командного зачёта; `event_id` — ID зачёта. Передавайте `result_id`, `competition_id` и `discipline_id` из одной записи истории в заявку. `place`, `score`, `result_status` и `protocol_url` могут быть `null`. Пример получения:

```bash
curl 'http://localhost:8080/api/v1/profile/competition-results?discipline_id=1' \
  -H "Authorization: Bearer $ATHLETE_TOKEN"
```

```json
{
  "athlete_user_id": 17,
  "rank_id": 3,
  "discipline_id": 1,
  "application_type": "assignment",
  "competition_id": 8,
  "result_id": 25
}
```

Например, подача за себя:

```bash
curl -X POST http://localhost:8080/api/v1/rank-applications \
  -H "Authorization: Bearer $ATHLETE_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"rank_id":3,"discipline_id":1,"application_type":"assignment","competition_id":8,"result_id":25}'
```

`athlete_user_id` можно не передавать при подаче за себя. Если по дисциплине у спортсмена есть подтверждённый тренер, заявку подаёт этот тренер. Backend проверяет принадлежность результата спортсмену, дисциплине и соревнованию, статус `finished`, одобренную заявку на соревнование и утверждённый зачёт с соответствующей записью результата. Наличие подходящего результата не гарантирует присвоение: при одобрении дополнительно проверяется норматив ЕВСК.

История возвращает `401` без токена и `400` при неверном `discipline_id`. Создание заявки возвращает `400` при неверном или неподтверждённом основании, `403` при отсутствии прав и `409` при дубликате. При рассмотрении одобрение без выполненного норматива ЕВСК возвращает `422`.

### Просмотр

- `GET /api/v1/rank-applications/mine` — собственные заявки и их статусы;
- `GET /api/v1/ranks/history` — история подтверждённых присвоений со сроком, приказом, соревнованием и результатом;
- `GET /api/v1/admin/rank-applications?limit=50&offset=0` — очередь федерации, требуется `rank_applications.review`.

Пустые списки возвращаются как `data: []`.

### Рассмотрение

`PATCH /api/v1/admin/rank-applications/{applicationID}` требует `rank_applications.review`.

Одобрение:

```json
{
  "status": "approved",
  "assigned_at": "2026-10-27",
  "expires_at": "2028-10-27",
  "order_number": "105-р"
}
```

Отклонение:

```json
{
  "status": "rejected",
  "rejection_reason": "Результат не соответствует заявленному нормативу"
}
```

`assigned_at` обязателен для одобрения. `expires_at` необязателен, но не может быть раньше даты присвоения. Для отклонения обязательна причина.

## Контроль сроков

### Миграция

`017_report_deadlines.up.sql` добавляет минимальную таблицу `report_deadline`: название, срок, статус и автор. Статусы: `planned`, `completed`, `cancelled`. Планировщик рассылок не добавлялся.

### Ролевой список

`GET /api/v1/deadlines` требует access-токен и формирует ответ по ролям и подтверждённым связям:

- `participant` — сроки регистрации доступных соревнований и собственных разрядов;
- `coach` — сроки регистрации и сроки разрядов подтверждённых подопечных по общей дисциплине;
- `judge` — срок собственной категории и ближайшие назначения;
- пользователь с `deadlines.manage` — все регистрации, разряды, судейские категории и отчётные сроки.

Элементы сортируются по `due_at`, затем стабильно по типу и ID. Пользователь с несколькими ролями не получает дубликаты.

```json
{
  "data": [
    {
      "type": "rank_expiration",
      "id": 19,
      "title": "Кандидат в мастера спорта",
      "due_at": "2027-05-20T00:00:00Z",
      "status": "active",
      "user_id": 17,
      "user_name": "Иванов Иван"
    }
  ]
}
```

### Отчётные сроки федерации

- `GET /api/v1/admin/report-deadlines`;
- `POST /api/v1/admin/report-deadlines`;
- `PATCH /api/v1/admin/report-deadlines/{id}`.

Требуется `deadlines.manage`.

```json
{
  "name": "Ходатайства на губернаторские премии",
  "due_at": "2027-02-01",
  "status": "planned"
}
```

Дата принимает `YYYY-MM-DD` или RFC 3339. PATCH обновляет только переданные поля; пустой объект запрещён. Плановые будущие отчётные сроки также попадают в дашборд федерации.

## Интеграция Competition Event

В текущей ветке пакет `internal/competitionevent` и функция `registerCompetitionEventRoutes` отсутствуют. Они находятся в отдельной ветке `features/competition_event`. Код из неё не переносился и ветки не объединялись.

После штатной интеграции Event-ветки в `cmd/api/main.go` нужно инициализировать её Repository, Service и Handler и вызвать существующую `registerCompetitionEventRoutes(mux, competitionEventHandler, protected)` рядом с остальными Competition routes. До интеграции добавлять один вызов нельзя: текущая ветка не скомпилируется.
