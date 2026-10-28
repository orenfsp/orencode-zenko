-- 1. Уведомление при публикации результата
CREATE OR REPLACE FUNCTION notify_competition_result()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    athlete_id bigint;
    competition_name text;
    message_text text;
BEGIN
    SELECT athlete_user_id INTO athlete_id
    FROM competition_application
    WHERE id = NEW.application_id;

    IF athlete_id IS NOT NULL AND NEW.place IS NOT NULL THEN
        SELECT c.name INTO competition_name
        FROM competition c
        JOIN competition_event e ON e.competition_id = c.id
        JOIN competition_application a ON a.event_id = e.id
        WHERE a.id = NEW.application_id;

        message_text := 'Опубликованы результаты соревнования «' || COALESCE(competition_name, '') || '». Ваше место: ' || NEW.place::text;

        INSERT INTO notification_outbox (recipient_id, payload)
        VALUES (
            athlete_id,
            jsonb_build_object(
                'type', 'competition_result_published',
                'application_id', NEW.application_id,
                'place', NEW.place,
                'text', message_text
            )
        );
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_competition_result_notification
AFTER INSERT OR UPDATE OF place ON competition_result
FOR EACH ROW
EXECUTE FUNCTION notify_competition_result();

-- 2. Уведомление при открытии регистрации
CREATE OR REPLACE FUNCTION notify_competition_registration_opened()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    athlete_id bigint;
    message_text text;
BEGIN
    IF NEW.status = 'registration_open' AND (OLD.status IS NULL OR OLD.status != 'registration_open') THEN
        message_text := 'Открыта регистрация на соревнование «' || NEW.name || '». Успейте подать заявку!';

        FOR athlete_id IN
            SELECT DISTINCT ru.user_id
            FROM role_user ru
            JOIN role r ON r.id = ru.role_id
            WHERE r.code = 'participant'
        LOOP
            INSERT INTO notification_outbox (recipient_id, payload)
            VALUES (
                athlete_id,
                jsonb_build_object(
                    'type', 'competition_registration_opened',
                    'competition_id', NEW.id,
                    'text', message_text
                )
            );
        END LOOP;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_competition_registration_notification
AFTER UPDATE OF status ON competition
FOR EACH ROW
EXECUTE FUNCTION notify_competition_registration_opened();

-- 3. Уведомление при подаче заявки на разряд
CREATE OR REPLACE FUNCTION notify_rank_application_created()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    admin_user_id bigint;
    rank_name text;
    message_text text;
BEGIN
    SELECT r.name INTO rank_name FROM rank r WHERE r.id = NEW.rank_id;
    
    INSERT INTO notification_outbox (recipient_id, payload)
    VALUES (
        NEW.athlete_user_id,
        jsonb_build_object(
            'type', 'rank_application_created',
            'application_id', NEW.id,
            'rank_id', NEW.rank_id,
            'text', 'Ваша заявка на присвоение разряда «' || COALESCE(rank_name, '') || '» подана и находится на рассмотрении.'
        )
    );

    FOR admin_user_id IN
        SELECT DISTINCT ru.user_id
        FROM role_user ru
        JOIN role_permission rp ON rp.role_id = ru.role_id
        JOIN permission p ON p.id = rp.permission_id
        JOIN "user" u ON u.id = ru.user_id
        WHERE p.code = 'rank_applications.review'
          AND u.access_status = true
    LOOP
        INSERT INTO notification_outbox (recipient_id, payload)
        VALUES (
            admin_user_id,
            jsonb_build_object(
                'type', 'rank_application_review_needed',
                'application_id', NEW.id,
                'athlete_user_id', NEW.athlete_user_id,
                'text', 'Новая заявка на присвоение разряда от спортсмена (ID: ' || NEW.athlete_user_id || ').'
            )
        );
    END LOOP;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_rank_application_notification
AFTER INSERT ON rank_application
FOR EACH ROW
EXECUTE FUNCTION notify_rank_application_created();