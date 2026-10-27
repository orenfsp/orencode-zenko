-- name: JudgeCategoryExists :one
SELECT EXISTS (SELECT 1 FROM judge_category WHERE id = sqlc.arg(category_id));

-- name: CreateJudgeCategoryApplication :one
INSERT INTO judge_category_application (judge_user_id, category_id, application_type)
VALUES (sqlc.arg(judge_user_id), sqlc.arg(category_id), sqlc.arg(application_type))
RETURNING id, judge_user_id, category_id, application_type, status, submitted_at,
          reviewed_by_user_id, reviewed_at, assigned_at, expires_at, order_number, rejection_reason;

-- name: ListJudgeCategoryApplicationsByJudge :many
SELECT
    jca.id, jca.judge_user_id, jca.category_id, jc.name AS category_name,
    jca.application_type, jca.status, jca.submitted_at,
    jca.reviewed_by_user_id, jca.reviewed_at, jca.assigned_at, jca.expires_at,
    jca.order_number, jca.rejection_reason
FROM judge_category_application AS jca
JOIN judge_category AS jc ON jc.id = jca.category_id
WHERE jca.judge_user_id = sqlc.arg(judge_user_id)
ORDER BY jca.submitted_at DESC, jca.id DESC;

-- name: ListPendingJudgeCategoryApplications :many
SELECT
    jca.id, jca.judge_user_id, u.last_name, u.first_name, u.middle_name,
    jca.category_id, jc.name AS category_name, jca.application_type,
    jca.status, jca.submitted_at, jca.reviewed_by_user_id, jca.reviewed_at,
    jca.assigned_at, jca.expires_at, jca.order_number, jca.rejection_reason
FROM judge_category_application AS jca
JOIN "user" AS u ON u.id = jca.judge_user_id
JOIN judge_category AS jc ON jc.id = jca.category_id
WHERE jca.status = 'pending'
ORDER BY jca.submitted_at, jca.id
LIMIT sqlc.arg(p_limit) OFFSET sqlc.arg(p_offset);

-- name: GetJudgeCategoryApplicationForUpdate :one
SELECT id, judge_user_id, category_id, application_type, status, submitted_at,
       reviewed_by_user_id, reviewed_at, assigned_at, expires_at, order_number, rejection_reason
FROM judge_category_application
WHERE id = sqlc.arg(id)
FOR UPDATE;

-- name: ReviewJudgeCategoryApplication :one
UPDATE judge_category_application
SET status = sqlc.arg(status),
    reviewed_by_user_id = sqlc.arg(reviewed_by_user_id),
    reviewed_at = now(),
    assigned_at = sqlc.narg(assigned_at),
    expires_at = sqlc.narg(expires_at),
    order_number = sqlc.narg(order_number),
    rejection_reason = sqlc.narg(rejection_reason)
WHERE id = sqlc.arg(id) AND status = 'pending'
RETURNING id, judge_user_id, category_id, application_type, status, submitted_at,
          reviewed_by_user_id, reviewed_at, assigned_at, expires_at, order_number, rejection_reason;

-- name: CreateUserJudgeCategoryFromApplication :one
INSERT INTO user_judge_category (user_id, category_id, assigned_at, expires_at, order_number)
VALUES (sqlc.arg(user_id), sqlc.arg(category_id), sqlc.arg(assigned_at), sqlc.narg(expires_at), sqlc.narg(order_number))
RETURNING id;

-- name: GetCurrentJudgeCategory :one
SELECT
    ujc.id, ujc.user_id, ujc.category_id, jc.name AS category_name,
    ujc.assigned_at, ujc.expires_at, ujc.order_number
FROM user_judge_category AS ujc
JOIN judge_category AS jc ON jc.id = ujc.category_id
WHERE ujc.user_id = sqlc.arg(user_id)
  AND (ujc.expires_at IS NULL OR ujc.expires_at >= CURRENT_DATE)
ORDER BY ujc.assigned_at DESC, ujc.id DESC
LIMIT 1;

-- name: ListJudgeCompetitionAssignments :many
SELECT
    cj.id, cj.competition_id, c.name AS competition_name,
    c.start_at, c.end_at, c.status,
    cj.position_id, jp.name AS position_name, cj.appointed_at
FROM competition_judge AS cj
JOIN competition AS c ON c.id = cj.competition_id
JOIN judge_position AS jp ON jp.id = cj.position_id
WHERE cj.judge_user_id = sqlc.arg(judge_user_id)
ORDER BY c.start_at DESC, c.id DESC, cj.id DESC;

