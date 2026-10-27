-- name: GetAntiDopingFileForAttachment :one
SELECT *
FROM uploaded_file
WHERE id = sqlc.arg(file_id)
FOR UPDATE;

-- name: CreateAntiDopingCertificate :one
INSERT INTO anti_doping (user_id, file_id, status)
VALUES (sqlc.arg(user_id), sqlc.arg(file_id), 'manual_review')
RETURNING *;

-- name: ListAntiDopingCertificatesByUser :many
SELECT *
FROM anti_doping
WHERE user_id = sqlc.arg(user_id)
ORDER BY created_at DESC, id DESC;

-- name: GetAntiDopingCertificateForReview :one
SELECT *
FROM anti_doping
WHERE id = sqlc.arg(certificate_id)
FOR UPDATE;

-- name: ApproveAntiDopingCertificate :one
UPDATE anti_doping
SET status = CASE
        WHEN make_date(EXTRACT(YEAR FROM sqlc.arg(completed_at)::date)::integer, 12, 31) < CURRENT_DATE
            THEN 'expired'
        ELSE 'valid'
    END,
    completed_at = sqlc.arg(completed_at)::date,
    valid_until = make_date(EXTRACT(YEAR FROM sqlc.arg(completed_at)::date)::integer, 12, 31),
    reviewed_by_user_id = sqlc.arg(reviewed_by_user_id),
    reviewed_at = now(),
    rejection_reason = NULL,
    updated_at = now()
WHERE id = sqlc.arg(certificate_id)
  AND status = 'manual_review'
  AND sqlc.arg(completed_at)::date <= CURRENT_DATE
RETURNING *;

-- name: RejectAntiDopingCertificate :one
UPDATE anti_doping
SET status = 'rejected',
    completed_at = NULL,
    valid_until = NULL,
    reviewed_by_user_id = sqlc.arg(reviewed_by_user_id),
    reviewed_at = now(),
    rejection_reason = sqlc.arg(rejection_reason),
    updated_at = now()
WHERE id = sqlc.arg(certificate_id)
  AND status = 'manual_review'
RETURNING *;

-- name: ExpireAntiDopingCertificates :execrows
UPDATE anti_doping
SET status = 'expired',
    updated_at = now()
WHERE status = 'valid'
  AND valid_until < CURRENT_DATE;
