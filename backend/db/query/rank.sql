-- name: CreateRank :one

INSERT INTO rank (name, ttl) VALUES ($1, $2) RETURNING *;

-- name: GetRank :one
SELECT * FROM rank WHERE id = $1;

-- name: GetListRank :many
SELECT * FROM rank ORDER BY created_at DESC;
