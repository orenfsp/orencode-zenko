-- name: UpsertEVSKStandard :one
INSERT INTO rank_evsk_standard (rank_id, discipline_id, minimum_place, minimum_score, valid_from, valid_to)
VALUES (sqlc.arg(rank_id), sqlc.arg(discipline_id), sqlc.narg(minimum_place), sqlc.narg(minimum_score)::numeric, sqlc.narg(valid_from), sqlc.narg(valid_to))
ON CONFLICT (rank_id, discipline_id) DO UPDATE SET
    minimum_place = EXCLUDED.minimum_place,
    minimum_score = EXCLUDED.minimum_score,
    valid_from = EXCLUDED.valid_from,
    valid_to = EXCLUDED.valid_to
RETURNING id, rank_id, discipline_id, minimum_place, minimum_score::text AS minimum_score, valid_from, valid_to, created_at;

-- name: ListEVSKStandards :many
SELECT id, rank_id, discipline_id, minimum_place, minimum_score::text AS minimum_score, valid_from, valid_to, created_at
FROM rank_evsk_standard
ORDER BY rank_id, discipline_id;
