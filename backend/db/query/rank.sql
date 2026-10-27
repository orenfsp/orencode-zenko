-- name: CreateRank :one

INSERT INTO rank (name, ttl) VALUES ($1, $2) RETURNING *;

-- name: GetRank :one
SELECT * FROM rank WHERE id = $1;

-- name: GetListRank :many
SELECT * FROM rank ORDER BY created_at DESC;

-- name: DeleteExpiredUserRanks :execrows
DELETE FROM user_rank AS user_rank
USING rank
WHERE rank.id = user_rank.rank_id
  AND rank.ttl IS NOT NULL
  AND user_rank.created_at + make_interval(years => rank.ttl) <= now();

-- name: DeleteExpiredUserRankDisciplines :execrows
DELETE FROM user_rank_discipline AS user_rank_discipline
USING rank
WHERE rank.id = user_rank_discipline.rank_id
  AND rank.ttl IS NOT NULL
  AND user_rank_discipline.created_at + make_interval(years => rank.ttl) <= now();
