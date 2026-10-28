-- name: CreateResultTemplate :one
INSERT INTO discipline_result_template (discipline_id, name, fields)
VALUES (sqlc.arg(discipline_id), sqlc.arg(name), sqlc.arg(fields)) RETURNING *;

-- name: GetResultTemplate :one
SELECT * FROM discipline_result_template WHERE id = sqlc.arg(id);

-- name: ListResultTemplates :many
SELECT * FROM discipline_result_template
WHERE discipline_id = sqlc.arg(discipline_id)
ORDER BY name, id;

-- name: UpdateResultTemplate :one
UPDATE discipline_result_template
SET name = sqlc.arg(name), fields = sqlc.arg(fields), is_active = sqlc.arg(is_active), updated_at = now()
WHERE id = sqlc.arg(id) RETURNING *;
