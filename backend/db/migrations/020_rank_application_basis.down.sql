DROP INDEX IF EXISTS rank_application_result_idx;

ALTER TABLE rank_application
    DROP CONSTRAINT IF EXISTS rank_application_dates_check,
    DROP CONSTRAINT IF EXISTS rank_application_type_check,
    DROP COLUMN IF EXISTS order_number,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS assigned_at,
    DROP COLUMN IF EXISTS result_id,
    DROP COLUMN IF EXISTS competition_id,
    DROP COLUMN IF EXISTS application_type;

CREATE OR REPLACE FUNCTION apply_rank_application_status() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS NOT DISTINCT FROM NEW.status THEN
        RETURN NEW;
    END IF;

    IF NEW.status = 'approved' THEN
        INSERT INTO user_rank_discipline (user_id, rank_id, discipline_id, awarded_at)
        SELECT NEW.athlete_user_id, NEW.rank_id, NEW.discipline_id, CURRENT_DATE
        WHERE NOT EXISTS (
            SELECT 1 FROM user_rank_discipline
            WHERE user_id = NEW.athlete_user_id
              AND rank_id = NEW.rank_id
              AND discipline_id = NEW.discipline_id
              AND result_id IS NULL
        );
    END IF;

    IF NEW.status IN ('approved', 'rejected') THEN
        INSERT INTO notification_outbox (recipient_id, payload)
        VALUES (
            NEW.athlete_user_id,
            jsonb_build_object(
                'type', 'rank_application_status_changed',
                'rank_application_id', NEW.id,
                'status', NEW.status,
                'text', CASE NEW.status
                    WHEN 'approved' THEN 'Ваша заявка на получение разряда одобрена.'
                    ELSE 'Ваша заявка на получение разряда отклонена.'
                END
            )
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

