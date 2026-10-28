-- name: ListDeadlineRoleCodes :many
SELECT r.code
FROM role_user AS ru
JOIN role AS r ON r.id = ru.role_id
WHERE ru.user_id = sqlc.arg(user_id)
ORDER BY r.code;

-- name: ListRegistrationDeadlines :many
SELECT c.id, c.name, c.registration_ends_at AS due_at, c.status
FROM competition AS c
WHERE c.registration_ends_at >= now()
  AND c.status IN ('published', 'registration_open')
ORDER BY c.registration_ends_at, c.id;

-- name: ListAthleteRankDeadlines :many
SELECT urd.id, urd.user_id, u.last_name, u.first_name, u.middle_name,
       r.name AS subject_name, urd.expires_at AS due_at
FROM user_rank_discipline AS urd
JOIN "user" AS u ON u.id = urd.user_id
JOIN rank AS r ON r.id = urd.rank_id
WHERE urd.user_id = sqlc.arg(user_id)
  AND urd.expires_at >= CURRENT_DATE
ORDER BY urd.expires_at, urd.id;

-- name: ListCoachAthleteRankDeadlines :many
SELECT urd.id, urd.user_id, u.last_name, u.first_name, u.middle_name,
       r.name AS subject_name, urd.expires_at AS due_at
FROM user_coach AS uc
JOIN user_rank_discipline AS urd
  ON urd.user_id = uc.user_id AND urd.discipline_id = uc.discipline_id
JOIN "user" AS u ON u.id = urd.user_id
JOIN rank AS r ON r.id = urd.rank_id
WHERE uc.coach_user_id = sqlc.arg(coach_user_id)
  AND uc.status = 'confirmed'
  AND urd.expires_at >= CURRENT_DATE
ORDER BY urd.expires_at, urd.id;

-- name: ListAllRankDeadlines :many
SELECT urd.id, urd.user_id, u.last_name, u.first_name, u.middle_name,
       r.name AS subject_name, urd.expires_at AS due_at
FROM user_rank_discipline AS urd
JOIN "user" AS u ON u.id = urd.user_id
JOIN rank AS r ON r.id = urd.rank_id
WHERE urd.expires_at >= CURRENT_DATE
ORDER BY urd.expires_at, urd.id;

-- name: ListJudgeCategoryDeadlines :many
SELECT ujc.id, ujc.user_id, u.last_name, u.first_name, u.middle_name,
       jc.name AS subject_name, ujc.expires_at AS due_at
FROM user_judge_category AS ujc
JOIN "user" AS u ON u.id = ujc.user_id
JOIN judge_category AS jc ON jc.id = ujc.category_id
WHERE ujc.user_id = sqlc.arg(user_id)
  AND ujc.expires_at >= CURRENT_DATE
ORDER BY ujc.expires_at, ujc.id;

-- name: ListAllJudgeCategoryDeadlines :many
SELECT ujc.id, ujc.user_id, u.last_name, u.first_name, u.middle_name,
       jc.name AS subject_name, ujc.expires_at AS due_at
FROM user_judge_category AS ujc
JOIN "user" AS u ON u.id = ujc.user_id
JOIN judge_category AS jc ON jc.id = ujc.category_id
WHERE ujc.expires_at >= CURRENT_DATE
ORDER BY ujc.expires_at, ujc.id;

-- name: ListJudgeAssignmentDeadlines :many
SELECT cj.id, c.id AS competition_id, c.name AS competition_name,
       c.start_at AS due_at, jp.name AS position_name
FROM competition_judge AS cj
JOIN competition AS c ON c.id = cj.competition_id
JOIN judge_position AS jp ON jp.id = cj.position_id
WHERE cj.judge_user_id = sqlc.arg(judge_user_id)
  AND c.start_at >= now()
ORDER BY c.start_at, c.id, cj.id;

-- name: ListAllJudgeAssignmentDeadlines :many
SELECT cj.id, c.id AS competition_id, c.name AS competition_name,
       c.start_at AS due_at, jp.name AS position_name
FROM competition_judge AS cj
JOIN competition AS c ON c.id = cj.competition_id
JOIN judge_position AS jp ON jp.id = cj.position_id
WHERE c.start_at >= now()
ORDER BY c.start_at, c.id, cj.id;

-- name: ListReportDeadlines :many
SELECT id, name, due_at, status, deadline_type, created_by_user_id, created_at, updated_at
FROM report_deadline
ORDER BY due_at, id;

-- name: CreateReportDeadline :one
INSERT INTO report_deadline (name, due_at, status, deadline_type, created_by_user_id)
VALUES (sqlc.arg(name), sqlc.arg(due_at), sqlc.arg(status), sqlc.arg(deadline_type), sqlc.arg(created_by_user_id))
RETURNING id, name, due_at, status, deadline_type, created_by_user_id, created_at, updated_at;

-- name: GetReportDeadline :one
SELECT id, name, due_at, status, deadline_type, created_by_user_id, created_at, updated_at
FROM report_deadline
WHERE id = sqlc.arg(id);

-- name: UpdateReportDeadline :one
UPDATE report_deadline
SET name = sqlc.arg(name), due_at = sqlc.arg(due_at), status = sqlc.arg(status), deadline_type = sqlc.arg(deadline_type), updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, name, due_at, status, deadline_type, created_by_user_id, created_at, updated_at;

-- name: CreateReportDeadlineReminders :exec
WITH candidates AS (
    SELECT d.id AS deadline_id, u.id AS recipient_id, sqlc.arg(days_before)::integer AS days_before,
           d.name, d.due_at, d.status
    FROM report_deadline AS d
    JOIN role_user AS ru ON ru.user_id IS NOT NULL
    JOIN role_permission AS rp ON rp.role_id = ru.role_id
    JOIN permission AS p ON p.id = rp.permission_id AND p.code = 'deadlines.manage'
    JOIN "user" AS u ON u.id = ru.user_id AND u.access_status = true
    WHERE d.status = 'planned'
      AND d.due_at >= now() + make_interval(days => sqlc.arg(days_before)::integer)
      AND d.due_at < now() + make_interval(days => sqlc.arg(days_before)::integer + 1)
), claimed AS (
    INSERT INTO report_deadline_reminder (deadline_id, recipient_id, days_before)
    SELECT deadline_id, recipient_id, days_before
    FROM candidates
    ON CONFLICT (deadline_id, recipient_id, days_before) DO NOTHING
    RETURNING deadline_id, recipient_id, days_before
)
INSERT INTO notification_outbox (recipient_id, payload)
SELECT c.recipient_id,
       jsonb_build_object(
           'type', 'report_deadline_reminder',
           'report_deadline_id', d.id,
           'days_before', c.days_before,
           'due_at', d.due_at,
           'text', format('Отчётный срок «%s» наступает через %s дн.', d.name, c.days_before)
       )
FROM claimed AS c
JOIN report_deadline AS d ON d.id = c.deadline_id;
