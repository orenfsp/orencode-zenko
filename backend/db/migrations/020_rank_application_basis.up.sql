ALTER TABLE rank_application
    ADD COLUMN application_type varchar(20) NOT NULL DEFAULT 'assignment',
    ADD COLUMN competition_id bigint REFERENCES competition (id),
    ADD COLUMN result_id bigint REFERENCES competition_result (id),
    ADD COLUMN assigned_at date,
    ADD COLUMN expires_at date,
    ADD COLUMN order_number varchar(100),
    ADD CONSTRAINT rank_application_type_check
        CHECK (application_type IN ('assignment', 'confirmation')),
    ADD CONSTRAINT rank_application_dates_check
        CHECK (expires_at IS NULL OR assigned_at IS NULL OR expires_at >= assigned_at);

CREATE INDEX rank_application_result_idx ON rank_application (result_id) WHERE result_id IS NOT NULL;

CREATE OR REPLACE FUNCTION apply_rank_application_status() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS NOT DISTINCT FROM NEW.status THEN
        RETURN NEW;
    END IF;

    IF NEW.status = 'approved' THEN
        INSERT INTO user_rank_discipline (
            user_id, rank_id, discipline_id, result_id, awarded_at, expires_at, order_number
        ) VALUES (
            NEW.athlete_user_id,
            NEW.rank_id,
            NEW.discipline_id,
            NEW.result_id,
            COALESCE(NEW.assigned_at, CURRENT_DATE),
            NEW.expires_at,
            NEW.order_number
        )
        ON CONFLICT (result_id) WHERE result_id IS NOT NULL
        DO UPDATE SET
            user_id = EXCLUDED.user_id,
            rank_id = EXCLUDED.rank_id,
            discipline_id = EXCLUDED.discipline_id,
            awarded_at = EXCLUDED.awarded_at,
            expires_at = EXCLUDED.expires_at,
            order_number = EXCLUDED.order_number;
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

