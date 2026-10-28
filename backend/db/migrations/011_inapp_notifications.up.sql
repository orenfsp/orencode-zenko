CREATE TABLE inapp_notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Индекс для быстрого поиска непрочитанных уведомлений пользователя
CREATE INDEX idx_inapp_notifications_user_unread ON inapp_notifications (user_id) WHERE is_read = FALSE;