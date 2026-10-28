CREATE OR REPLACE FUNCTION notify_outbox() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('notification_events', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_outbox_notify ON notification_outbox;
CREATE TRIGGER trg_outbox_notify
AFTER INSERT ON notification_outbox
FOR EACH ROW EXECUTE FUNCTION notify_outbox();