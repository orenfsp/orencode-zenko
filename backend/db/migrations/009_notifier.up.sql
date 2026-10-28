CREATE TABLE notification_outbox (
    id BIGSERIAL PRIMARY KEY,
    recipient_id BIGINT NOT NULL,
    payload JSONB NOT NULL,
    telegram_sent BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE telegram_links (
    user_id BIGINT PRIMARY KEY,
    chat_id BIGINT NOT NULL,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION notify_outbox() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('notification_events', NEW.id::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_outbox_notify
AFTER INSERT ON notification_outbox
FOR EACH ROW EXECUTE FUNCTION notify_outbox();