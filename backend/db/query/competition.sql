-- name: ListPublishedCompetitions :many
SELECT
    c.id,
    c.name,
    c.description,
    c.discipline_id,
    d.name AS discipline_name,
    c.competition_type_id,
    ct.name AS competition_type_name,
    c.competition_level_id,
    cl.code AS competition_level_code,
    cl.name AS competition_level_name,
    c.format,
    c.start_at,
    c.end_at,
    c.registration_starts_at,
    c.registration_ends_at,
    c.venue_address,
    c.online_url,
    c.status,
    c.calendar_plan_id,
    cp.name AS calendar_plan_name,
    cp.year AS calendar_plan_year,
    c.calendar_plan_number,
    c.regulation_url
FROM competition AS c
JOIN discipline AS d ON d.id = c.discipline_id
JOIN competition_type AS ct ON ct.id = c.competition_type_id
JOIN competition_level AS cl ON cl.id = c.competition_level_id
LEFT JOIN calendar_plan AS cp ON cp.id = c.calendar_plan_id
WHERE c.status = 'published'
ORDER BY c.start_at, c.id
LIMIT $1 OFFSET $2;

-- name: GetPublishedCompetition :one
SELECT
    c.id,
    c.name,
    c.description,
    c.discipline_id,
    d.name AS discipline_name,
    c.competition_type_id,
    ct.name AS competition_type_name,
    c.competition_level_id,
    cl.code AS competition_level_code,
    cl.name AS competition_level_name,
    c.format,
    c.start_at,
    c.end_at,
    c.registration_starts_at,
    c.registration_ends_at,
    c.venue_address,
    c.online_url,
    c.status,
    c.calendar_plan_id,
    cp.name AS calendar_plan_name,
    cp.year AS calendar_plan_year,
    c.calendar_plan_number,
    c.regulation_url
FROM competition AS c
JOIN discipline AS d ON d.id = c.discipline_id
JOIN competition_type AS ct ON ct.id = c.competition_type_id
JOIN competition_level AS cl ON cl.id = c.competition_level_id
LEFT JOIN calendar_plan AS cp ON cp.id = c.calendar_plan_id
WHERE c.id = $1 AND c.status = 'published';

-- name: ValidateCompetitionReferences :one
SELECT
    EXISTS(SELECT 1 FROM discipline WHERE discipline.id = sqlc.arg(p_discipline_id)) AS discipline_exists,
    EXISTS(SELECT 1 FROM competition_type WHERE competition_type.id = sqlc.arg(p_competition_type_id)) AS competition_type_exists,
    EXISTS(SELECT 1 FROM competition_level WHERE competition_level.id = sqlc.arg(p_competition_level_id)) AS competition_level_exists,
    (
        sqlc.narg(p_calendar_plan_id)::bigint IS NULL
        OR EXISTS(SELECT 1 FROM calendar_plan WHERE calendar_plan.id = sqlc.narg(p_calendar_plan_id))
    ) AS calendar_plan_exists;

-- name: CreateCompetition :one
WITH created AS (
    INSERT INTO competition (
        name,
        description,
        discipline_id,
        competition_type_id,
        competition_level_id,
        format,
        start_at,
        end_at,
        registration_starts_at,
        registration_ends_at,
        venue_address,
        online_url,
        created_by,
        status,
        calendar_plan_id,
        calendar_plan_number,
        regulation_url
    ) VALUES (
        sqlc.arg(name),
        sqlc.narg(description),
        sqlc.arg(discipline_id),
        sqlc.arg(competition_type_id),
        sqlc.arg(competition_level_id),
        sqlc.arg(format),
        sqlc.arg(start_at),
        sqlc.arg(end_at),
        sqlc.arg(registration_starts_at),
        sqlc.arg(registration_ends_at),
        sqlc.narg(venue_address),
        sqlc.narg(online_url),
        sqlc.arg(created_by),
        'draft',
        sqlc.narg(calendar_plan_id),
        sqlc.narg(calendar_plan_number),
        sqlc.narg(regulation_url)
    )
    RETURNING
        id,
        name,
        description,
        discipline_id,
        competition_type_id,
        competition_level_id,
        format,
        start_at,
        end_at,
        registration_starts_at,
        registration_ends_at,
        venue_address,
        online_url,
        status,
        calendar_plan_id,
        calendar_plan_number,
        regulation_url
)
SELECT
    c.id,
    c.name,
    c.description,
    c.discipline_id,
    d.name AS discipline_name,
    c.competition_type_id,
    ct.name AS competition_type_name,
    c.competition_level_id,
    cl.code AS competition_level_code,
    cl.name AS competition_level_name,
    c.format,
    c.start_at,
    c.end_at,
    c.registration_starts_at,
    c.registration_ends_at,
    c.venue_address,
    c.online_url,
    c.status,
    c.calendar_plan_id,
    cp.name AS calendar_plan_name,
    cp.year AS calendar_plan_year,
    c.calendar_plan_number,
    c.regulation_url
FROM created AS c
JOIN discipline AS d ON d.id = c.discipline_id
JOIN competition_type AS ct ON ct.id = c.competition_type_id
JOIN competition_level AS cl ON cl.id = c.competition_level_id
LEFT JOIN calendar_plan AS cp ON cp.id = c.calendar_plan_id;
