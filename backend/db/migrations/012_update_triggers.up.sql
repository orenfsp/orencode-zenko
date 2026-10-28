CREATE OR REPLACE FUNCTION notify_outbox() RETURNS TRIGGER AS $$
DECLARE
    msg_text TEXT;
BEGIN
    -- 1. Отправляем сигнал для Telegram-бота
    PERFORM pg_notify('notification_events', NEW.id::text);

    -- 2. Параллельно создаем уведомление для фронта (In-App)
    -- Достаем текст из JSONB payload
    msg_text := NEW.payload->>'text';
    
    IF msg_text IS NULL THEN
        msg_text := 'Новое уведомление';
    END IF;

    INSERT INTO inapp_notifications (user_id, title, message)
    VALUES (NEW.recipient_id, 'Уведомление', msg_text);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Пересоздаем триггер (удаляем старый и вешаем новый)
DROP TRIGGER IF EXISTS trg_outbox_notify ON notification_outbox;
CREATE TRIGGER trg_outbox_notify
AFTER INSERT ON notification_outbox
FOR EACH ROW EXECUTE FUNCTION notify_outbox();