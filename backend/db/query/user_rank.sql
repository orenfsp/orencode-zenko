-- name: CreateUserCoach :one
INSERT INTO user_coach (
    user_id,
    coach_user_id,
    discipline_id
)
VALUES (
    sqlc.arg(participant_user_id),
    sqlc.arg(coach_user_id),
    sqlc.arg(discipline_id)
)
RETURNING *;


-- name: ListUserCoachesByParticipant :many
SELECT *
FROM user_coach
WHERE user_id = sqlc.arg(participant_user_id)
ORDER BY requested_at DESC, id DESC
LIMIT sqlc.arg(page_limit)::integer
OFFSET sqlc.arg(page_offset)::integer;


-- name: ListUserCoachesByCoach :many
SELECT *
FROM user_coach
WHERE coach_user_id = sqlc.arg(coach_user_id)
ORDER BY requested_at DESC, id DESC
LIMIT sqlc.arg(page_limit)::integer
OFFSET sqlc.arg(page_offset)::integer;


-- name: ConfirmUserCoach :one
UPDATE user_coach
SET
    status = 'confirmed',
    confirmed_at = now(),
    ended_at = NULL
WHERE id = sqlc.arg(relation_id)
  AND user_id = sqlc.arg(participant_user_id)
  AND status = 'pending'
RETURNING *;


-- name: EndUserCoach :one
UPDATE user_coach
SET
    status = 'ended',
    ended_at = now()
WHERE id = sqlc.arg(relation_id)
  AND (
      user_id = sqlc.arg(actor_user_id)
      OR coach_user_id = sqlc.arg(actor_user_id)
  )
  AND status IN ('pending', 'confirmed')
RETURNING *;