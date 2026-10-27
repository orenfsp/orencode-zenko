-- name: CreateAuditLog :exec
INSERT INTO audit_log (
    event_type,
    actor_user_id,
    entity_type,
    entity_id,
    method,
    route,
    status
) VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListAuditLogs :many
SELECT id, event_type, actor_user_id, entity_type, entity_id, method, route, status, occurred_at
FROM audit_log
ORDER BY occurred_at DESC, id DESC
LIMIT $1 OFFSET $2;
