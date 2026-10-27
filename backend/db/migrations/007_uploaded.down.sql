ALTER TABLE "parental_consent"
    DROP CONSTRAINT IF EXISTS "parental_consent_file_id_fkey",
    DROP COLUMN IF EXISTS "file_id";

ALTER TABLE "consent_data"
    DROP CONSTRAINT IF EXISTS "consent_data_file_id_fkey",
    DROP COLUMN IF EXISTS "file_id";

DROP TABLE IF EXISTS "uploaded_file";
