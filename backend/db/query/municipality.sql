
-- name: CreateMunicipality :one
INSERT INTO municipality (name)
VALUES ($1)
RETURNING id, name;

-- name: GetMunicipalityByID :one
SELECT id, name
FROM municipality
WHERE id = $1;

-- name: ListMunicipalities :many
SELECT id, name
FROM municipality
ORDER BY name
LIMIT $1 OFFSET $2;
