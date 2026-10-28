ALTER TABLE "competition_event"
    ADD COLUMN "name" varchar(250);

UPDATE "competition_event"
SET "name" = 'Зачёт №' || "id"
WHERE "name" IS NULL;

ALTER TABLE "competition_event"
    ALTER COLUMN "name" SET NOT NULL;
