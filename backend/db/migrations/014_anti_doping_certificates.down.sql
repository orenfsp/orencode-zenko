DROP TRIGGER IF EXISTS "competition_application_valid_anti_doping_trigger" ON "competition_application";
DROP FUNCTION IF EXISTS require_valid_anti_doping_certificate();
DROP INDEX IF EXISTS "anti_doping_user_valid_idx";

ALTER TABLE "anti_doping"
    DROP CONSTRAINT IF EXISTS "anti_doping_review_data_check",
    DROP CONSTRAINT IF EXISTS "anti_doping_status_check",
    DROP CONSTRAINT IF EXISTS "anti_doping_reviewed_by_user_id_fkey",
    DROP CONSTRAINT IF EXISTS "anti_doping_file_id_key",
    DROP CONSTRAINT IF EXISTS "anti_doping_file_id_fkey",
    DROP COLUMN IF EXISTS "rejection_reason",
    DROP COLUMN IF EXISTS "reviewed_at",
    DROP COLUMN IF EXISTS "reviewed_by_user_id",
    DROP COLUMN IF EXISTS "valid_until",
    DROP COLUMN IF EXISTS "status",
    DROP COLUMN IF EXISTS "file_id";
