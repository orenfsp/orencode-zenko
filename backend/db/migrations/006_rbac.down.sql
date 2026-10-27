DROP INDEX IF EXISTS "role_permission_permission_id_idx";
DROP INDEX IF EXISTS "role_user_user_id_idx";
DROP TABLE IF EXISTS "role_permission";
DROP TABLE IF EXISTS "permission";

ALTER TABLE "role"
    DROP CONSTRAINT IF EXISTS "role_code_key",
    DROP COLUMN IF EXISTS "self_assignable",
    DROP COLUMN IF EXISTS "is_system",
    DROP COLUMN IF EXISTS "code";
