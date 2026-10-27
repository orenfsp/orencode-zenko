-- name: CreateAuthUser :one
INSERT INTO "user" (
    first_name,
    last_name,
    middle_name,
    birth,
    email,
    password_hash
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, email, password_hash, tg_username, access_status;

-- name: CreateConsentData :exec
INSERT INTO consent_data (user_id, file_id)
VALUES ($1, $2);

-- name: CreateParentalConsent :exec
INSERT INTO parental_consent (user_id, file_id)
VALUES ($1, $2);

-- name: GetAuthUserByEmail :one
SELECT
    u.id,
    u.email,
    u.password_hash,
    u.tg_username,
    u.access_status,
    COALESCE(selected_role.role_id, 0)::bigint AS role_id
FROM "user" u
LEFT JOIN LATERAL (
    SELECT ru.role_id
    FROM role_user ru
    WHERE ru.user_id = u.id
    ORDER BY ru.granted_at DESC, ru.id DESC
    LIMIT 1
) selected_role ON true
WHERE u.email = $1;

-- name: GetAuthUserByID :one
SELECT
    u.id,
    u.email,
    u.password_hash,
    u.tg_username,
    u.access_status,
    COALESCE(selected_role.role_id, 0)::bigint AS role_id
FROM "user" u
LEFT JOIN LATERAL (
    SELECT ru.role_id
    FROM role_user ru
    WHERE ru.user_id = u.id
    ORDER BY ru.granted_at DESC, ru.id DESC
    LIMIT 1
) selected_role ON true
WHERE u.id = $1;

-- name: IsRoleSelfAssignable :one
SELECT EXISTS (
    SELECT 1
    FROM role
    WHERE id = $1 AND self_assignable = true
);

-- name: DeleteUserRoles :exec
DELETE FROM role_user
WHERE user_id = $1;

-- name: UpdateTelegramUsername :execrows
UPDATE "user"
SET tg_username = $2, updated_at = now()
WHERE id = $1;

-- name: CreateUserSession :one
INSERT INTO user_session (user_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, refresh_token_hash, expires_at;

-- name: GetUserSessionByRefreshHash :one
SELECT id, user_id, refresh_token_hash, expires_at
FROM user_session
WHERE refresh_token_hash = $1;

-- name: RotateUserSession :execrows
UPDATE user_session
SET refresh_token_hash = $2, expires_at = $3, updated_at = now()
WHERE id = $1;

-- name: DeleteUserSession :execrows
DELETE FROM user_session
WHERE id = $1;
