-- name: ListDisciplines :many
SELECT id, name
FROM discipline
ORDER BY name ASC, id ASC;

-- name: ListCompetitionTypes :many
SELECT id, name
FROM competition_type
ORDER BY name ASC, id ASC;

-- name: ListCompetitionLevels :many
SELECT id, code, name
FROM competition_level
ORDER BY name ASC NULLS LAST, code ASC NULLS LAST, id ASC;

-- name: ListAgeGroups :many
SELECT id, name, min_age, max_age
FROM age_group
ORDER BY min_age ASC NULLS LAST, max_age ASC NULLS LAST, name ASC, id ASC;

-- name: UpsertDiscipline :exec
INSERT INTO discipline (name)
VALUES (sqlc.arg(name))
ON CONFLICT (name) DO NOTHING;
