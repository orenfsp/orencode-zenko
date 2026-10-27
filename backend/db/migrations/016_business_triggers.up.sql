CREATE OR REPLACE FUNCTION notify_competition_application()
RETURNS TRIGGER AS $$
DECLARE
    msg_text TEXT;
BEGIN
    -- Реагируем только если статус реально изменился
    IF NEW.status IS DISTINCT FROM OLD.status THEN
        IF NEW.status = 'approved' THEN
            msg_text := 'Ваша заявка на соревнование одобрена!';
        ELSIF NEW.status = 'rejected' THEN
            msg_text := 'Ваша заявка на соревнование отклонена.';
        ELSE
            -- На случай других статусов (например, 'cancelled')
            msg_text := 'Статус вашей заявки изменился на: ' || NEW.status;
        END IF;

        -- Кладем событие в очередь для Telegram и фронта
        INSERT INTO notification_outbox (recipient_id, payload)
        VALUES (NEW.athlete_user_id, jsonb_build_object('text', msg_text));
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_competition_application_status
AFTER UPDATE ON competition_application
FOR EACH ROW
EXECUTE FUNCTION notify_competition_application();