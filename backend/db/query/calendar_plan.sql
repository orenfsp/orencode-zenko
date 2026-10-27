-- name: ListCalendarPlans :many
SELECT id, name, year
FROM calendar_plan
WHERE sqlc.narg(p_year)::smallint IS NULL
   OR year = sqlc.narg(p_year)
ORDER BY year DESC, id ASC
LIMIT sqlc.arg(p_limit)
OFFSET sqlc.arg(p_offset);

-- name: GetCalendarPlan :one
SELECT id, name, year
FROM calendar_plan
WHERE id = sqlc.arg(p_id);

-- name: CreateCalendarPlan :one
INSERT INTO calendar_plan (name, year)
VALUES (sqlc.arg(p_name), sqlc.arg(p_year))
RETURNING id, name, year;

-- name: UpdateCalendarPlan :one
UPDATE calendar_plan
SET name = sqlc.arg(p_name),
    year = sqlc.arg(p_year)
WHERE id = sqlc.arg(p_id)
RETURNING id, name, year;

-- name: ListCalendarPlanCompetitions :many
SELECT
    c.id,
    c.name,
    c.start_at,
    c.end_at,
    c.status,
    c.discipline_id,
    d.name AS discipline_name,
    c.calendar_plan_number
FROM competition AS c
JOIN discipline AS d ON d.id = c.discipline_id
WHERE c.calendar_plan_id = sqlc.arg(p_calendar_plan_id)::bigint
ORDER BY c.start_at ASC, c.id ASC
LIMIT sqlc.arg(p_limit)
OFFSET sqlc.arg(p_offset);
