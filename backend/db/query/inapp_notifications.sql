-- name: GetUserNotifications :many
SELECT * FROM inapp_notifications 
WHERE user_id = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: MarkNotificationAsRead :exec
UPDATE inapp_notifications 
SET is_read = TRUE 
WHERE id = $1 AND user_id = $2;

-- name: CreateInAppNotification :exec
INSERT INTO inapp_notifications (user_id, title, message) 
VALUES ($1, $2, $3);