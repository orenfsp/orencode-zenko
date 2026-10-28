-- name: GetCompetitionApplicationEventContext :one
SELECT
    ce.id AS event_id,
    ce.discipline_id,
    c.id AS competition_id,
    c.status AS competition_status,
    c.registration_starts_at,
    c.registration_ends_at
FROM competition_event AS ce
JOIN competition AS c ON c.id = ce.competition_id
WHERE ce.id = sqlc.arg(event_id);

-- name: IsConfirmedCoachForCompetitionApplication :one
SELECT EXISTS (
    SELECT 1
    FROM user_coach
    WHERE user_id = sqlc.arg(athlete_user_id)
      AND coach_user_id = sqlc.arg(coach_user_id)
      AND discipline_id = sqlc.arg(discipline_id)
      AND status = 'confirmed'
      AND ended_at IS NULL
);

-- name: CreateCompetitionApplication :one
INSERT INTO competition_application (
    event_id,
    athlete_user_id,
    submitted_by_user_id,
    status
)
VALUES (
    sqlc.arg(event_id),
    sqlc.arg(athlete_user_id),
    sqlc.arg(submitted_by_user_id),
    'pending'
)
RETURNING *;

-- name: GetCompetitionApplication :one
SELECT
    ca.id,
    ca.event_id,
    ca.athlete_user_id,
    ca.submitted_by_user_id,
    ca.status,
    ca.submitted_at,
    ca.reviewed_by_user_id,
    ca.reviewed_at,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    c.name AS competition_name,
    c.start_at AS competition_start_at,
    c.end_at AS competition_end_at,
    c.registration_starts_at,
    c.registration_ends_at,
    athlete.first_name AS athlete_first_name,
    athlete.last_name AS athlete_last_name,
    athlete.middle_name AS athlete_middle_name,
    athlete.email AS athlete_email,
    submitter.first_name AS submitter_first_name,
    submitter.last_name AS submitter_last_name,
    submitter.middle_name AS submitter_middle_name
FROM competition_application AS ca
JOIN competition_event AS ce ON ce.id = ca.event_id
JOIN competition AS c ON c.id = ce.competition_id
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
JOIN "user" AS athlete ON athlete.id = ca.athlete_user_id
JOIN "user" AS submitter ON submitter.id = ca.submitted_by_user_id
WHERE ca.id = sqlc.arg(application_id);

-- name: ListAccessibleCompetitionApplications :many
SELECT
    ca.id,
    ca.event_id,
    ca.athlete_user_id,
    ca.submitted_by_user_id,
    ca.status,
    ca.submitted_at,
    ca.reviewed_by_user_id,
    ca.reviewed_at,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    c.name AS competition_name,
    c.start_at AS competition_start_at,
    c.end_at AS competition_end_at,
    c.registration_starts_at,
    c.registration_ends_at,
    athlete.first_name AS athlete_first_name,
    athlete.last_name AS athlete_last_name,
    athlete.middle_name AS athlete_middle_name,
    athlete.email AS athlete_email,
    submitter.first_name AS submitter_first_name,
    submitter.last_name AS submitter_last_name,
    submitter.middle_name AS submitter_middle_name
FROM competition_application AS ca
JOIN competition_event AS ce ON ce.id = ca.event_id
JOIN competition AS c ON c.id = ce.competition_id
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
JOIN "user" AS athlete ON athlete.id = ca.athlete_user_id
JOIN "user" AS submitter ON submitter.id = ca.submitted_by_user_id
WHERE ca.athlete_user_id = sqlc.arg(actor_user_id)
   OR EXISTS (
       SELECT 1
       FROM user_coach AS uc
       WHERE uc.user_id = ca.athlete_user_id
         AND uc.coach_user_id = sqlc.arg(actor_user_id)
         AND uc.discipline_id = ce.discipline_id
         AND uc.status = 'confirmed'
         AND uc.ended_at IS NULL
   )
ORDER BY ca.submitted_at DESC, ca.id DESC;

-- name: ListCompetitionApplicationsForReview :many
SELECT
    ca.id,
    ca.event_id,
    ca.athlete_user_id,
    ca.submitted_by_user_id,
    ca.status,
    ca.submitted_at,
    ca.reviewed_by_user_id,
    ca.reviewed_at,
    ce.competition_id,
    ce.discipline_id,
    d.name AS discipline_name,
    ce.age_group_id,
    ag.name AS age_group_name,
    c.name AS competition_name,
    c.start_at AS competition_start_at,
    c.end_at AS competition_end_at,
    c.registration_starts_at,
    c.registration_ends_at,
    athlete.first_name AS athlete_first_name,
    athlete.last_name AS athlete_last_name,
    athlete.middle_name AS athlete_middle_name,
    athlete.email AS athlete_email,
    submitter.first_name AS submitter_first_name,
    submitter.last_name AS submitter_last_name,
    submitter.middle_name AS submitter_middle_name
FROM competition_application AS ca
JOIN competition_event AS ce ON ce.id = ca.event_id
JOIN competition AS c ON c.id = ce.competition_id
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
JOIN "user" AS athlete ON athlete.id = ca.athlete_user_id
JOIN "user" AS submitter ON submitter.id = ca.submitted_by_user_id
WHERE (sqlc.narg(p_event_id)::bigint IS NULL OR ca.event_id = sqlc.narg(p_event_id))
  AND (sqlc.narg(p_status)::text IS NULL OR ca.status = sqlc.narg(p_status))
ORDER BY ca.submitted_at DESC, ca.id DESC
LIMIT sqlc.arg(p_limit) OFFSET sqlc.arg(p_offset);

-- name: ReviewCompetitionApplication :one
UPDATE competition_application
SET
    status = sqlc.arg(status),
    reviewed_by_user_id = sqlc.arg(reviewed_by_user_id),
    reviewed_at = now()
WHERE id = sqlc.arg(application_id)
  AND status = 'pending'
RETURNING *;

-- name: WithdrawCompetitionApplication :one
UPDATE competition_application AS ca
SET status = 'withdrawn'
FROM competition_event AS ce
JOIN competition AS c ON c.id = ce.competition_id
WHERE ca.id = sqlc.arg(application_id)
  AND ca.event_id = ce.id
  AND ca.athlete_user_id = sqlc.arg(athlete_user_id)
  AND ca.status = 'pending'
  AND c.registration_ends_at >= now()
RETURNING ca.*;
