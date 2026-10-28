-- name: GetIntegrationClientByKey :one
SELECT ic.id, ic.client_key, ic.name, ic.secret_hash, ic.owner_user_id, ic.is_active
FROM integration_client AS ic
JOIN "user" AS owner ON owner.id = ic.owner_user_id
WHERE ic.client_key = sqlc.arg(client_key)
  AND ic.is_active = true
  AND owner.access_status = true;

-- name: GetIntegrationClientByID :one
SELECT ic.id, ic.client_key, ic.name, ic.owner_user_id, ic.is_active
FROM integration_client AS ic
JOIN "user" AS owner ON owner.id = ic.owner_user_id
WHERE ic.id = sqlc.arg(client_id)
  AND ic.is_active = true
  AND owner.access_status = true;

-- name: ListIntegrationClientPermissions :many
SELECT permission_code
FROM integration_client_permission
WHERE client_id = sqlc.arg(client_id)
ORDER BY permission_code;

-- name: IntegrationClientHasPermission :one
SELECT EXISTS (
    SELECT 1
    FROM integration_client AS ic
    JOIN "user" AS owner ON owner.id = ic.owner_user_id
    JOIN integration_client_permission AS icp ON icp.client_id = ic.id
    WHERE ic.id = sqlc.arg(client_id)
      AND ic.is_active = true
      AND owner.access_status = true
      AND icp.permission_code = sqlc.arg(permission_code)
);

-- name: GetIntegrationOwnerIDByEmail :one
SELECT id FROM "user" WHERE lower(email) = lower(sqlc.arg(email)) AND access_status = true;

-- name: UpsertIntegrationClient :one
INSERT INTO integration_client (client_key, name, secret_hash, owner_user_id, is_active)
VALUES (sqlc.arg(client_key), sqlc.arg(name), sqlc.arg(secret_hash), sqlc.arg(owner_user_id), true)
ON CONFLICT (client_key) DO UPDATE
SET name = EXCLUDED.name,
    secret_hash = EXCLUDED.secret_hash,
    owner_user_id = EXCLUDED.owner_user_id,
    is_active = true,
    updated_at = now()
RETURNING id;

-- name: ClearIntegrationClientPermissions :exec
DELETE FROM integration_client_permission WHERE client_id = sqlc.arg(client_id);

-- name: GrantIntegrationClientPermission :exec
INSERT INTO integration_client_permission (client_id, permission_code)
VALUES (sqlc.arg(client_id), sqlc.arg(permission_code))
ON CONFLICT DO NOTHING;

-- name: CreateWebhookSubscription :one
INSERT INTO webhook_subscription (client_id, url, signing_secret)
VALUES (sqlc.arg(client_id), sqlc.arg(url), sqlc.arg(signing_secret))
RETURNING id, client_id, url, is_active, created_at;

-- name: AddWebhookSubscriptionEvent :exec
INSERT INTO webhook_subscription_event (subscription_id, event_type)
VALUES (sqlc.arg(subscription_id), sqlc.arg(event_type));

-- name: ListWebhookSubscriptions :many
SELECT id, client_id, url, is_active, created_at
FROM webhook_subscription
WHERE client_id = sqlc.arg(client_id)
ORDER BY id;

-- name: ListWebhookSubscriptionEvents :many
SELECT event_type
FROM webhook_subscription_event
WHERE subscription_id = sqlc.arg(subscription_id)
ORDER BY event_type;

-- name: DeleteWebhookSubscription :execrows
DELETE FROM webhook_subscription
WHERE id = sqlc.arg(subscription_id) AND client_id = sqlc.arg(client_id);

-- name: EnqueueWebhookDeliveries :exec
INSERT INTO webhook_delivery (subscription_id, event_type, payload)
SELECT ws.id, sqlc.arg(event_type), sqlc.arg(payload)::jsonb
FROM webhook_subscription AS ws
JOIN webhook_subscription_event AS wse ON wse.subscription_id = ws.id
JOIN integration_client AS ic ON ic.id = ws.client_id
WHERE ws.is_active = true
  AND ic.is_active = true
  AND wse.event_type = sqlc.arg(event_type);

-- name: ClaimWebhookDeliveries :many
WITH claimed AS (
    SELECT wd.id
    FROM webhook_delivery AS wd
    WHERE (
        (wd.status = 'pending' AND wd.next_attempt_at <= now())
        OR (wd.status = 'processing' AND wd.updated_at < now() - interval '5 minutes')
    )
    ORDER BY wd.next_attempt_at, wd.id
    FOR UPDATE SKIP LOCKED
    LIMIT sqlc.arg(batch_size)
)
UPDATE webhook_delivery AS wd
SET status = 'processing', attempts = attempts + 1, updated_at = now()
FROM claimed, webhook_subscription AS ws
WHERE wd.id = claimed.id AND ws.id = wd.subscription_id
RETURNING wd.id, wd.event_type, wd.payload, wd.attempts, ws.url, ws.signing_secret, wd.created_at;

-- name: MarkWebhookDeliverySent :exec
UPDATE webhook_delivery
SET status = 'sent', response_status = sqlc.arg(response_status), last_error = NULL,
    sent_at = now(), updated_at = now()
WHERE id = sqlc.arg(delivery_id);

-- name: RetryWebhookDelivery :exec
UPDATE webhook_delivery
SET status = sqlc.arg(status), response_status = sqlc.narg(response_status),
    last_error = sqlc.arg(last_error), next_attempt_at = sqlc.arg(next_attempt_at), updated_at = now()
WHERE id = sqlc.arg(delivery_id);
