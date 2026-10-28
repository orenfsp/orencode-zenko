DROP TRIGGER IF EXISTS trg_competition_result_rank ON competition_result;
DROP FUNCTION IF EXISTS apply_rank_from_competition_result();
DROP TRIGGER IF EXISTS trg_rank_application_status ON rank_application;
DROP FUNCTION IF EXISTS apply_rank_application_status();
DROP INDEX IF EXISTS user_rank_discipline_result_id_idx;
ALTER TABLE competition_result DROP COLUMN IF EXISTS awarded_rank_id;
DROP TABLE IF EXISTS rank_application;
