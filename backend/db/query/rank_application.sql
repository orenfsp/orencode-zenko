-- name: HasConfirmedCoachForRankApplication :one
SELECT EXISTS (
    SELECT 1
    FROM user_coach
    WHERE user_id = sqlc.arg(athlete_user_id)
      AND discipline_id = sqlc.arg(discipline_id)
      AND status = 'confirmed'
);

-- name: IsConfirmedCoachForRankApplication :one
SELECT EXISTS (
    SELECT 1
    FROM user_coach
    WHERE user_id = sqlc.arg(athlete_user_id)
      AND coach_user_id = sqlc.arg(coach_user_id)
      AND discipline_id = sqlc.arg(discipline_id)
      AND status = 'confirmed'
);

-- name: CreateRankApplication :one
INSERT INTO rank_application (
    athlete_user_id,
    rank_id,
    discipline_id,
    submitted_by_user_id,
    application_type,
    competition_id,
    result_id
)
VALUES (
    sqlc.arg(athlete_user_id),
    sqlc.arg(rank_id),
    sqlc.arg(discipline_id),
    sqlc.arg(submitted_by_user_id),
    sqlc.arg(application_type),
    sqlc.arg(competition_id),
    sqlc.arg(result_id)
)
RETURNING *;

-- name: ValidateRankApplicationBasis :one
SELECT
    EXISTS (SELECT 1 FROM competition_result AS basis WHERE basis.id = sqlc.arg(result_id)) AS result_exists,
    EXISTS (
        SELECT 1
        FROM competition_result AS cr
        JOIN competition_application AS ca ON ca.id = cr.application_id
        WHERE cr.id = sqlc.arg(result_id)
          AND ca.athlete_user_id = sqlc.arg(athlete_user_id)
    ) AS athlete_matches,
    EXISTS (
        SELECT 1
        FROM competition_result AS cr
        JOIN competition_application AS ca ON ca.id = cr.application_id
        JOIN competition_event AS ce ON ce.id = ca.event_id
        WHERE cr.id = sqlc.arg(result_id)
          AND ce.discipline_id = sqlc.arg(discipline_id)
    ) AS discipline_matches,
    EXISTS (
        SELECT 1
        FROM competition_result AS cr
        JOIN competition_application AS ca ON ca.id = cr.application_id
        JOIN competition_event AS ce ON ce.id = ca.event_id
        WHERE cr.id = sqlc.arg(result_id)
          AND ce.competition_id = sqlc.arg(competition_id)
    ) AS competition_matches,
    EXISTS (
        SELECT 1 FROM competition_result AS basis
        WHERE basis.id = sqlc.arg(result_id) AND basis.result_status = 'finished'
    ) AS result_finished;

-- name: ReviewRankApplication :one
UPDATE rank_application
SET
    status = sqlc.arg(status),
    reviewed_by_user_id = sqlc.arg(reviewed_by_user_id),
    reviewed_at = now(),
    rejection_reason = sqlc.narg(rejection_reason),
    assigned_at = sqlc.narg(assigned_at),
    expires_at = sqlc.narg(expires_at),
    order_number = sqlc.narg(order_number),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'pending'
RETURNING *;

-- name: ListRankApplicationsByAthlete :many
SELECT *
FROM rank_application
WHERE athlete_user_id = sqlc.arg(athlete_user_id)
ORDER BY created_at DESC, id DESC;

-- name: ListPendingRankApplications :many
SELECT *
FROM rank_application
WHERE status = 'pending'
ORDER BY created_at, id
LIMIT sqlc.arg(p_limit) OFFSET sqlc.arg(p_offset);

-- name: ListUserRankHistory :many
SELECT
    urd.id, urd.user_id, urd.rank_id, r.name AS rank_name,
    urd.discipline_id, d.name AS discipline_name,
    urd.result_id, urd.awarded_at, urd.expires_at, urd.order_number,
    c.id AS competition_id, c.name AS competition_name
FROM user_rank_discipline AS urd
JOIN rank AS r ON r.id = urd.rank_id
JOIN discipline AS d ON d.id = urd.discipline_id
LEFT JOIN competition_result AS cr ON cr.id = urd.result_id
LEFT JOIN competition_application AS ca ON ca.id = cr.application_id
LEFT JOIN competition_event AS ce ON ce.id = ca.event_id
LEFT JOIN competition AS c ON c.id = ce.competition_id
WHERE urd.user_id = sqlc.arg(user_id)
ORDER BY urd.awarded_at DESC NULLS LAST, urd.id DESC;
