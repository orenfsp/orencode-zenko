-- name: CreateUploadedFile :one
INSERT INTO uploaded_file (
    purpose_code,
    storage_key,
    original_filename,
    content_type,
    size_bytes,
    sha256
)
VALUES (
    sqlc.arg(purpose_code),
    sqlc.arg(storage_key),
    sqlc.arg(original_filename),
    sqlc.arg(content_type),
    sqlc.arg(size_bytes),
    sqlc.arg(sha256)
) RETURNING *;


-- name: DeleteUploadedFileByStorageKey :execrows
DELETE FROM uploaded_file
WHERE storage_key = sqlc.arg(storage_key)
  AND status = 'uploaded';


-- name: GetUploadedFileStatusByStorageKey :one
SELECT status
FROM uploaded_file
WHERE storage_key = sqlc.arg(storage_key);


-- name: GetUploadedFileByID :one
SELECT *
FROM uploaded_file
WHERE id = sqlc.arg(file_id);


-- name: GetUploadedFileForAttachment :one
SELECT *
FROM uploaded_file
WHERE id = sqlc.arg(file_id)
FOR UPDATE;


-- name: GetAttachedFileForDownload :one
SELECT
    uploaded_file.*,
    COALESCE(consent_data.user_id, parental_consent.user_id, 0)::bigint AS owner_user_id
FROM uploaded_file
LEFT JOIN consent_data ON consent_data.file_id = uploaded_file.id
LEFT JOIN parental_consent ON parental_consent.file_id = uploaded_file.id
WHERE uploaded_file.id = sqlc.arg(file_id)
  AND uploaded_file.status = 'attached';


-- name: MarkUploadedFileAttached :execrows
UPDATE uploaded_file
SET status = 'attached',
    updated_at = now()
WHERE id = sqlc.arg(file_id)
  AND status = 'uploaded';


-- name: MarkUploadedFileDeletePending :execrows
UPDATE uploaded_file
SET status = 'delete_pending',
    updated_at = now()
WHERE id = sqlc.arg(file_id)
  AND status IN ('uploaded', 'delete_pending');


-- name: MarkUploadedFileDeletePendingByStorageKey :execrows
UPDATE uploaded_file
SET status = 'delete_pending',
    updated_at = now()
WHERE storage_key = sqlc.arg(storage_key)
  AND status IN ('uploaded', 'delete_pending');


-- name: DeleteDiscardableUploadedFile :execrows
DELETE FROM uploaded_file
WHERE id = sqlc.arg(file_id)
  AND status IN ('uploaded', 'delete_pending');


-- name: ListFilesPendingCleanup :many
SELECT *
FROM uploaded_file
WHERE status = 'delete_pending'
   OR (status = 'uploaded' AND created_at < sqlc.arg(uploaded_before))
ORDER BY created_at
LIMIT sqlc.arg(batch_size);
