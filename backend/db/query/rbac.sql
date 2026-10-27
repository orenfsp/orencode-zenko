-- name: HasPermission :one
SELECT EXISTS (
    SELECT 1
    FROM "user" u
    JOIN role_user ru ON ru.user_id = u.id
    JOIN role_permission rp ON rp.role_id = ru.role_id
    JOIN permission p ON p.id = rp.permission_id
    WHERE u.id = $1
      AND u.access_status = true
      AND p.code = $2
);

-- name: HasRole :one
SELECT EXISTS (
    SELECT 1
    FROM role_user
    JOIN role ON role.id = role_user.role_id
    WHERE role_user.user_id = sqlc.arg(user_id)
      AND role.code = sqlc.arg(code)
);

-- name: CreateRole :one
INSERT INTO role (code, name, self_assignable)
VALUES ($1, $2, $3)
RETURNING id, code, name, is_system, self_assignable;

-- name: GetPermissionIDByCode :one
SELECT id
FROM permission
WHERE code = $1;

-- name: AssignRolePermission :exec
INSERT INTO role_permission (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- name: UserExists :one
SELECT EXISTS (
    SELECT 1
    FROM "user"
    WHERE id = $1
);

-- name: RoleExists :one
SELECT EXISTS (
    SELECT 1
    FROM role
    WHERE id = $1
);

-- name: AssignRole :exec
INSERT INTO role_user (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: ListRoles :many
SELECT id, code, name, is_system, self_assignable
FROM role
ORDER BY id;

-- name: ListPermissionCodesByRole :many
SELECT p.code
FROM permission p
JOIN role_permission rp ON rp.permission_id = p.id
WHERE rp.role_id = $1
ORDER BY p.code;

-- name: ListPermissions :many
SELECT id, code, description
FROM permission
ORDER BY code;

-- name: UpsertPermission :one
INSERT INTO permission (code, description)
VALUES ($1, $2)
ON CONFLICT (code) DO UPDATE
SET description = EXCLUDED.description
RETURNING id;

-- name: UpsertSuperAdminRole :one
INSERT INTO role (code, name, is_system, self_assignable)
VALUES ($1, $2, true, false)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    is_system = true,
    self_assignable = false,
    updated_at = now()
RETURNING id;

-- name: UpsertSelfAssignableRole :one
INSERT INTO role (code, name, is_system, self_assignable)
VALUES ($1, $2, false, true)
ON CONFLICT (code) DO UPDATE
SET name = EXCLUDED.name,
    is_system = false,
    self_assignable = true,
    updated_at = now()
RETURNING id;

-- name: UpsertSeedAdminUser :one
INSERT INTO "user" (
    first_name,
    last_name,
    birth,
    email,
    password_hash,
    access_status
) VALUES ($1, $2, $3, $4, $5, true)
ON CONFLICT (email) DO UPDATE
SET access_status = true,
    updated_at = now()
RETURNING id;

-- name: ListSelfAssignableRoles :many
SELECT id, code, name, is_system, self_assignable
FROM role
WHERE self_assignable = true AND is_system = false
ORDER BY name;
