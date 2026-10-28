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

-- name: ValidateRankApplicationEVSK :one
SELECT EXISTS (
    SELECT 1
    FROM rank_evsk_standard AS standard
    JOIN rank_application AS application ON application.id = sqlc.arg(application_id)
    JOIN competition_result AS cr ON cr.id = application.result_id
    JOIN competition_application AS ca ON ca.id = cr.application_id
    JOIN competition_event AS ce ON ce.id = ca.event_id
    JOIN competition AS c ON c.id = ce.competition_id
    WHERE standard.rank_id = application.rank_id
      AND standard.discipline_id = application.discipline_id
      AND cr.result_status = 'finished'
      AND ca.athlete_user_id = application.athlete_user_id
      AND ce.discipline_id = application.discipline_id
      AND (standard.valid_from IS NULL OR c.start_at::date >= standard.valid_from)
      AND (standard.valid_to IS NULL OR c.start_at::date <= standard.valid_to)
      AND (standard.minimum_place IS NULL OR cr.place IS NOT NULL AND cr.place <= standard.minimum_place)
      AND (standard.minimum_score IS NULL OR cr.score IS NOT NULL AND cr.score >= standard.minimum_score)
) AS valid;

-- name: ReviewRankApplication :one
UPDATE rank_application
SET
    status = sqlc.arg(status)::varchar,
    reviewed_by_user_id = sqlc.arg(reviewed_by_user_id),
    reviewed_at = now(),
    rejection_reason = sqlc.narg(rejection_reason),
    assigned_at = sqlc.narg(assigned_at),
    expires_at = COALESCE(
        sqlc.narg(expires_at),
        CASE
            WHEN sqlc.arg(status)::varchar = 'approved'::varchar AND rank.ttl IS NOT NULL
            THEN (COALESCE(sqlc.narg(assigned_at), CURRENT_DATE) + make_interval(years => rank.ttl))::date
            ELSE NULL
        END
    ),
    order_number = sqlc.narg(order_number),
    updated_at = now()
FROM rank
WHERE rank_application.id = sqlc.arg(id)
  AND rank_application.rank_id = rank.id
  AND rank_application.status = 'pending'
RETURNING rank_application.*;

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
