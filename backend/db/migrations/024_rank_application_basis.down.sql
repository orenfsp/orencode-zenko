DROP TRIGGER IF EXISTS trg_validate_rank_application_basis ON rank_application;
DROP FUNCTION IF EXISTS validate_rank_application_basis();

ALTER TABLE rank_application
    DROP CONSTRAINT IF EXISTS rank_application_basis_required_check;
