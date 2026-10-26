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
