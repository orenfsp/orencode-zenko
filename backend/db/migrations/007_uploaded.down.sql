UPDATE "consent_data" AS consent
SET "document_url" = 'internal-file:' || uploaded."storage_key"
FROM "uploaded_file" AS uploaded
WHERE consent."file_id" = uploaded."id"
  AND consent."document_url" IS NULL;

UPDATE "parental_consent" AS consent
SET "document_url" = 'internal-file:' || uploaded."storage_key"
FROM "uploaded_file" AS uploaded
WHERE consent."file_id" = uploaded."id"
  AND consent."document_url" IS NULL;

ALTER TABLE "parental_consent"
    DROP CONSTRAINT IF EXISTS "parental_consent_document_source_check",
    DROP CONSTRAINT IF EXISTS "parental_consent_file_id_fkey",
    DROP COLUMN IF EXISTS "file_id",
    ALTER COLUMN "document_url" SET NOT NULL;

ALTER TABLE "consent_data"
    DROP CONSTRAINT IF EXISTS "consent_data_document_source_check",
    DROP CONSTRAINT IF EXISTS "consent_data_file_id_fkey",
    DROP COLUMN IF EXISTS "file_id",
    ALTER COLUMN "document_url" SET NOT NULL;

DROP TABLE IF EXISTS "uploaded_file";
