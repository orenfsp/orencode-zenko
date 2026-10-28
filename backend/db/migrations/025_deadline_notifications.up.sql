CREATE OR REPLACE FUNCTION notify_report_deadline_change()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    admin_user_id bigint;
    event_type text;
    message text;
BEGIN
    IF TG_OP = 'INSERT' THEN
        event_type := 'report_deadline_created';
        message := format('Создан новый отчётный срок «%s» до %s.', NEW.name, to_char(NEW.due_at, 'YYYY-MM-DD'));
    ELSIF OLD.due_at IS DISTINCT FROM NEW.due_at
       OR OLD.status IS DISTINCT FROM NEW.status
       OR OLD.name IS DISTINCT FROM NEW.name THEN
        event_type := 'report_deadline_updated';
        message := format('Изменён отчётный срок «%s»: %s до %s.', NEW.name, NEW.status, to_char(NEW.due_at, 'YYYY-MM-DD'));
    ELSE
        RETURN NEW;
    END IF;

    FOR admin_user_id IN
        SELECT DISTINCT ru.user_id
        FROM role_user AS ru
        JOIN role_permission AS rp ON rp.role_id = ru.role_id
        JOIN permission AS p ON p.id = rp.permission_id
        JOIN "user" AS u ON u.id = ru.user_id
        WHERE p.code = 'deadlines.manage'
          AND u.access_status = true
    LOOP
        INSERT INTO notification_outbox (recipient_id, payload)
        VALUES (
            admin_user_id,
            jsonb_build_object(
                'type', event_type,
                'report_deadline_id', NEW.id,
                'status', NEW.status,
                'due_at', NEW.due_at,
                'text', message
            )
        );
    END LOOP;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_report_deadline_notification
AFTER INSERT OR UPDATE OF name, due_at, status ON report_deadline
FOR EACH ROW
EXECUTE FUNCTION notify_report_deadline_change();
