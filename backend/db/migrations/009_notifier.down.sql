DROP TRIGGER IF EXISTS trg_outbox_notify ON notification_outbox;
DROP FUNCTION IF EXISTS notify_outbox();
DROP TABLE IF EXISTS telegram_links;
DROP TABLE IF EXISTS notification_outbox;