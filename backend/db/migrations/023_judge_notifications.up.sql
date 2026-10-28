CREATE OR REPLACE FUNCTION notify_judge_category_application()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    category_name text;
    message_text text;
BEGIN
    SELECT name INTO category_name
    FROM judge_category
    WHERE id = NEW.category_id;

    IF TG_OP = 'INSERT' THEN
        message_text := 'Заявка на судейскую категорию «' || COALESCE(category_name, '') || '» принята на рассмотрение.';
        INSERT INTO notification_outbox (recipient_id, payload)
        VALUES (
            NEW.judge_user_id,
            jsonb_build_object(
                'type', 'judge_category_application_created',
                'application_id', NEW.id,
                'category_id', NEW.category_id,
                'application_type', NEW.application_type,
                'status', NEW.status,
                'text', message_text
            )
        );
        RETURN NEW;
    END IF;

    IF NEW.status IS DISTINCT FROM OLD.status
       AND NEW.status IN ('approved', 'rejected') THEN
        message_text := CASE NEW.status
            WHEN 'approved' THEN 'Заявка на судейскую категорию «' || COALESCE(category_name, '') || '» одобрена.'
            ELSE 'Заявка на судейскую категорию «' || COALESCE(category_name, '') || '» отклонена.'
        END;
        INSERT INTO notification_outbox (recipient_id, payload)
        VALUES (
            NEW.judge_user_id,
            jsonb_build_object(
                'type', 'judge_category_application_status_changed',
                'application_id', NEW.id,
                'category_id', NEW.category_id,
                'application_type', NEW.application_type,
                'status', NEW.status,
                'text', message_text
            )
        );
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_judge_category_application_notification
AFTER INSERT OR UPDATE OF status ON judge_category_application
FOR EACH ROW
EXECUTE FUNCTION notify_judge_category_application();

CREATE OR REPLACE FUNCTION notify_judge_assignment()
RETURNS trigger
LANGUAGE plpgsql
AS $$
DECLARE
    competition_name text;
    position_name text;
    message_text text;
    target_judge_id bigint;
    assignment_id bigint;
    competition_id bigint;
BEGIN
    IF TG_OP = 'INSERT' THEN
        target_judge_id := NEW.judge_user_id;
        assignment_id := NEW.id;
        competition_id := NEW.competition_id;
        SELECT c.name, jp.name
        INTO competition_name, position_name
        FROM competition AS c
        JOIN judge_position AS jp ON jp.id = NEW.position_id
        WHERE c.id = NEW.competition_id;
        message_text := 'Вы назначены на соревнование «' || COALESCE(competition_name, '') || '» (' || COALESCE(position_name, '') || ').';
    ELSE
        target_judge_id := OLD.judge_user_id;
        assignment_id := OLD.id;
        competition_id := OLD.competition_id;
        SELECT name INTO competition_name FROM competition WHERE id = OLD.competition_id;
        message_text := 'Ваше назначение на соревнование «' || COALESCE(competition_name, '') || '» снято.';
    END IF;

    INSERT INTO notification_outbox (recipient_id, payload)
    VALUES (
        target_judge_id,
        jsonb_build_object(
            'type', CASE WHEN TG_OP = 'INSERT' THEN 'judge_assignment_created' ELSE 'judge_assignment_removed' END,
            'assignment_id', assignment_id,
            'competition_id', competition_id,
            'text', message_text
        )
    );
    RETURN CASE WHEN TG_OP = 'INSERT' THEN NEW ELSE OLD END;
END;
$$;

CREATE TRIGGER trg_judge_assignment_notification
AFTER INSERT OR DELETE ON competition_judge
FOR EACH ROW
EXECUTE FUNCTION notify_judge_assignment();
