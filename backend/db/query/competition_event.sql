-- name: ListPublicCompetitionEvents :many
SELECT
    ce.id,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    ce.protocol_url
FROM competition_event AS ce
JOIN competition AS c ON c.id = ce.competition_id
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
WHERE ce.competition_id = sqlc.arg(competition_id)
  AND c.status = ANY(sqlc.arg(public_statuses)::text[])
ORDER BY ce.id;

-- name: GetPublicCompetitionEvent :one
SELECT
    ce.id,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    ce.protocol_url
FROM competition_event AS ce
JOIN competition AS c ON c.id = ce.competition_id
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
WHERE ce.competition_id = sqlc.arg(competition_id)
  AND ce.id = sqlc.arg(event_id)
  AND c.status = ANY(sqlc.arg(public_statuses)::text[]);

-- name: PublicCompetitionForEventsExists :one
SELECT EXISTS(
    SELECT 1
    FROM competition
    WHERE id = sqlc.arg(competition_id)
      AND status = ANY(sqlc.arg(public_statuses)::text[])
);

-- name: ListCompetitionEventsForAdmin :many
SELECT
    ce.id,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    ce.protocol_url
FROM competition_event AS ce
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
WHERE ce.competition_id = sqlc.arg(competition_id)
ORDER BY ce.id;

-- name: GetCompetitionEventForAdmin :one
SELECT
    ce.id,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    ce.protocol_url
FROM competition_event AS ce
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
WHERE ce.competition_id = sqlc.arg(competition_id)
  AND ce.id = sqlc.arg(event_id);

-- name: CompetitionForEventsExists :one
SELECT EXISTS(SELECT 1 FROM competition WHERE id = sqlc.arg(competition_id));

-- name: LockCompetitionForEventMutation :one
SELECT discipline_id, status
FROM competition
WHERE id = sqlc.arg(competition_id)
FOR UPDATE;

-- name: AgeGroupExists :one
SELECT EXISTS(SELECT 1 FROM age_group WHERE id = sqlc.arg(age_group_id));

-- name: LockCompetitionEventForMutation :one
SELECT
    ce.id,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    ce.protocol_url
FROM competition_event AS ce
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
WHERE ce.competition_id = sqlc.arg(competition_id)
  AND ce.id = sqlc.arg(event_id)
FOR UPDATE OF ce;

-- name: CompetitionEventHasApplications :one
SELECT EXISTS(
    SELECT 1
    FROM competition_application
    WHERE event_id = sqlc.arg(event_id)
);

-- name: CreateCompetitionEvent :one
WITH created AS (
    INSERT INTO competition_event (competition_id, discipline_id, age_group_id)
    VALUES (sqlc.arg(competition_id), sqlc.arg(discipline_id), sqlc.narg(age_group_id))
    RETURNING id, competition_id, discipline_id, age_group_id, protocol_url
)
SELECT
    ce.id,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    ce.protocol_url
FROM created AS ce
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id;

-- name: UpdateCompetitionEvent :one
WITH updated AS (
    UPDATE competition_event AS target
    SET age_group_id = sqlc.narg(age_group_id)
    WHERE target.competition_id = sqlc.arg(competition_id)
      AND target.id = sqlc.arg(event_id)
    RETURNING id, competition_id, discipline_id, age_group_id, protocol_url
)
SELECT
    ce.id,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    ce.protocol_url
FROM updated AS ce
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id;

-- name: DeleteCompetitionEvent :execrows
DELETE FROM competition_event
WHERE competition_id = sqlc.arg(competition_id)
  AND id = sqlc.arg(event_id);
