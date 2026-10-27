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
    submitted_by_user_id
)
VALUES (
    sqlc.arg(athlete_user_id),
    sqlc.arg(rank_id),
    sqlc.arg(discipline_id),
    sqlc.arg(submitted_by_user_id)
)
RETURNING *;

-- name: ReviewRankApplication :one
UPDATE rank_application
SET
    status = sqlc.arg(status),
    reviewed_by_user_id = sqlc.arg(reviewed_by_user_id),
    reviewed_at = now(),
    rejection_reason = sqlc.narg(rejection_reason),
    updated_at = now()
WHERE id = sqlc.arg(id)
  AND status = 'pending'
RETURNING *;

-- name: ListRankApplicationsByAthlete :many
SELECT *
FROM rank_application
WHERE athlete_user_id = sqlc.arg(athlete_user_id)
ORDER BY created_at DESC, id DESC;
