-- name: CreateTestTable :exec
INSERT INTO Test (id, username, password) VALUES ($1, $2, $3);