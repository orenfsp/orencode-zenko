DELETE FROM "audit_log"
WHERE "actor_user_id" IS NULL;

ALTER TABLE "audit_log"
    ALTER COLUMN "actor_user_id" SET NOT NULL;
