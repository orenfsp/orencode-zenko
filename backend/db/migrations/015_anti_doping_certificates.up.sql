TRUNCATE TABLE "anti_doping";

ALTER TABLE "anti_doping"
    ADD COLUMN "file_id" bigint NOT NULL,
    ADD COLUMN "status" varchar(20) NOT NULL DEFAULT 'manual_review',
    ADD COLUMN "valid_until" date,
    ADD COLUMN "reviewed_by_user_id" bigint,
    ADD COLUMN "reviewed_at" timestamptz,
    ADD COLUMN "rejection_reason" varchar(500),
    ADD CONSTRAINT "anti_doping_file_id_fkey"
        FOREIGN KEY ("file_id") REFERENCES "uploaded_file" ("id") ON DELETE RESTRICT,
    ADD CONSTRAINT "anti_doping_file_id_key" UNIQUE ("file_id"),
    ADD CONSTRAINT "anti_doping_reviewed_by_user_id_fkey"
        FOREIGN KEY ("reviewed_by_user_id") REFERENCES "user" ("id") ON DELETE RESTRICT,
    ADD CONSTRAINT "anti_doping_status_check"
        CHECK ("status" IN ('manual_review', 'valid', 'expired', 'rejected')),
    ADD CONSTRAINT "anti_doping_review_data_check" CHECK (
        ("status" = 'manual_review'
            AND "completed_at" IS NULL
            AND "valid_until" IS NULL
            AND "reviewed_by_user_id" IS NULL
            AND "reviewed_at" IS NULL
            AND "rejection_reason" IS NULL)
        OR
        ("status" IN ('valid', 'expired')
            AND "completed_at" IS NOT NULL
            AND "valid_until" = make_date(EXTRACT(YEAR FROM "completed_at")::integer, 12, 31)
            AND "reviewed_by_user_id" IS NOT NULL
            AND "reviewed_at" IS NOT NULL
            AND "rejection_reason" IS NULL)
        OR
        ("status" = 'rejected'
            AND "completed_at" IS NULL
            AND "valid_until" IS NULL
            AND "reviewed_by_user_id" IS NOT NULL
            AND "reviewed_at" IS NOT NULL
            AND length(btrim("rejection_reason")) BETWEEN 1 AND 500)
    );

CREATE INDEX "anti_doping_user_valid_idx"
    ON "anti_doping" ("user_id", "valid_until")
    WHERE "status" = 'valid';

CREATE OR REPLACE FUNCTION require_valid_anti_doping_certificate()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM anti_doping
        WHERE user_id = NEW.athlete_user_id
          AND status = 'valid'
          AND valid_until >= CURRENT_DATE
    ) THEN
        RAISE EXCEPTION 'athlete does not have a valid anti-doping certificate'
            USING ERRCODE = '23514',
                  CONSTRAINT = 'competition_application_valid_anti_doping_check';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER "competition_application_valid_anti_doping_trigger"
BEFORE INSERT ON "competition_application"
FOR EACH ROW
EXECUTE FUNCTION require_valid_anti_doping_certificate();
