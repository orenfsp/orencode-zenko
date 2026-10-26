ALTER TABLE IF EXISTS "user_rank_discipline"
    DROP CONSTRAINT IF EXISTS "user_rank_discipline_result_id_fkey";

DROP TABLE IF EXISTS "competition_result";
DROP TABLE IF EXISTS "competition_application";
