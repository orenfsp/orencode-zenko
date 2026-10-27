# Backend API пяти направлений MVP

Документ описывает только фактически реализованные backend-возможности. Он дополняется одновременно с каждым этапом MVP.

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

`athlete_user_id` можно не передавать при подаче за себя. Backend проверяет, что результат существует, имеет статус `finished`, относится к выбранным спортсмену, дисциплине и соревнованию. Норматив ЕВСК автоматически не оценивается: решение остаётся за федерацией.

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
