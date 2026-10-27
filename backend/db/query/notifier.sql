-- name: GetNotificationEvent :one
SELECT recipient_id, payload, telegram_sent
FROM notification_outbox
WHERE id = $1;

-- name: GetTelegramChatID :one
SELECT chat_id
FROM telegram_links
WHERE user_id = $1;

-- name: MarkTelegramSent :exec
UPDATE notification_outbox
SET telegram_sent = true
WHERE id = $1;