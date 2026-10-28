ALTER TABLE rank_application
    ADD CONSTRAINT rank_application_basis_required_check
    CHECK (competition_id IS NOT NULL AND result_id IS NOT NULL);

CREATE OR REPLACE FUNCTION validate_rank_application_basis()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    basis_athlete_id bigint;
    basis_discipline_id bigint;
    basis_competition_id bigint;
    basis_status varchar(30);
BEGIN
    SELECT ca.athlete_user_id, ce.discipline_id, ce.competition_id, cr.result_status
    INTO basis_athlete_id, basis_discipline_id, basis_competition_id, basis_status
    FROM competition_result AS cr
    JOIN competition_application AS ca ON ca.id = cr.application_id
    JOIN competition_event AS ce ON ce.id = ca.event_id
    WHERE cr.id = NEW.result_id;

    IF basis_athlete_id IS NULL
       OR basis_athlete_id <> NEW.athlete_user_id
       OR basis_discipline_id <> NEW.discipline_id
       OR basis_competition_id <> NEW.competition_id
       OR basis_status <> 'finished' THEN
        RAISE EXCEPTION 'rank application basis does not match a finished competition result'
            USING ERRCODE = '23514', CONSTRAINT = 'rank_application_basis_check';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_validate_rank_application_basis
BEFORE INSERT OR UPDATE OF athlete_user_id, discipline_id, competition_id, result_id
ON rank_application
FOR EACH ROW
EXECUTE FUNCTION validate_rank_application_basis();
