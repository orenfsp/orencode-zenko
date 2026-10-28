-- name: ListParentalConsents :many
SELECT pc.id, pc.user_id, pc.status, pc.submitted_at, pc.reviewed_at, pc.file_id,
       u.first_name, u.last_name, u.email
FROM parental_consent pc
JOIN "user" u ON u.id = pc.user_id
WHERE (sqlc.narg(status)::varchar IS NULL OR pc.status = sqlc.narg(status)::varchar)
ORDER BY pc.submitted_at DESC
LIMIT sqlc.arg(p_limit)::integer OFFSET sqlc.arg(p_offset)::integer;

-- name: GetParentalConsent :one
SELECT id, user_id, status, submitted_at, reviewed_at, file_id
FROM parental_consent
WHERE id = $1;

-- name: UpdateParentalConsentStatus :exec
UPDATE parental_consent
SET status = $2, reviewed_at = NOW()
WHERE id = $1;