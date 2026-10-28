-- Manual sheets cannot be represented by the original schema; refuse rollback
-- instead of silently deleting or changing their results.
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM competition_event_result_sheet WHERE scoring_mode = 'manual') THEN
        RAISE EXCEPTION 'cannot roll back result templates while manual sheets exist';
    END IF;
END $$;

ALTER TABLE competition_event_result_entry
    DROP CONSTRAINT result_entry_custom_values_check,
    DROP COLUMN custom_values;
ALTER TABLE competition_event_result_sheet
    DROP CONSTRAINT competition_event_result_sheet_task_count_check,
    ADD CONSTRAINT competition_event_result_sheet_task_count_check CHECK (task_count BETWEEN 1 AND 50),
    DROP CONSTRAINT result_sheet_template_check,
    DROP CONSTRAINT result_sheet_scoring_mode_check,
    DROP COLUMN template_snapshot,
    DROP COLUMN template_id,
    DROP COLUMN scoring_mode;
DROP TABLE discipline_result_template;
