-- name: GetFederationDashboardMetrics :one
SELECT
    (SELECT COUNT(DISTINCT ru.user_id)
     FROM role_user AS ru
     JOIN role AS r ON r.id = ru.role_id
     WHERE r.code = 'participant') AS athlete_count,
    (SELECT COUNT(*) FROM competition) AS competition_count,
    (SELECT COUNT(DISTINCT ca.athlete_user_id)
     FROM competition_application AS ca
     WHERE ca.status = 'approved') AS participant_count,
    (SELECT COUNT(*)
     FROM user_rank_discipline AS urd
     WHERE urd.expires_at IS NULL OR urd.expires_at >= CURRENT_DATE) AS active_rank_count;

-- name: ListDashboardUpcomingCompetitions :many
SELECT
    c.id,
    c.name,
    c.start_at,
    c.end_at,
    c.status,
    d.name AS discipline_name
FROM competition AS c
JOIN discipline AS d ON d.id = c.discipline_id
WHERE c.start_at >= now()
ORDER BY c.start_at, c.id
LIMIT sqlc.arg(p_limit);

-- name: ListDashboardPendingCompetitionApplications :many
SELECT
    ca.id,
    ca.submitted_at,
    u.id AS athlete_user_id,
    u.last_name,
    u.first_name,
    u.middle_name,
    c.id AS competition_id,
    c.name AS competition_name
FROM competition_application AS ca
JOIN "user" AS u ON u.id = ca.athlete_user_id
JOIN competition_event AS ce ON ce.id = ca.event_id
JOIN competition AS c ON c.id = ce.competition_id
WHERE ca.status = 'pending'
ORDER BY ca.submitted_at, ca.id
LIMIT sqlc.arg(p_limit);

-- name: ListDashboardPendingRankApplications :many
SELECT
    ra.id,
    ra.created_at,
    u.id AS athlete_user_id,
    u.last_name,
    u.first_name,
    u.middle_name,
    r.name AS rank_name,
    d.name AS discipline_name
FROM rank_application AS ra
JOIN "user" AS u ON u.id = ra.athlete_user_id
JOIN rank AS r ON r.id = ra.rank_id
JOIN discipline AS d ON d.id = ra.discipline_id
WHERE ra.status = 'pending'
ORDER BY ra.created_at, ra.id
LIMIT sqlc.arg(p_limit);

-- name: ListDashboardPendingJudgeCategoryApplications :many
SELECT
    jca.id,
    jca.submitted_at,
    u.id AS judge_user_id,
    u.last_name,
    u.first_name,
    u.middle_name,
    jc.name AS category_name,
    jca.application_type
FROM judge_category_application AS jca
JOIN "user" AS u ON u.id = jca.judge_user_id
JOIN judge_category AS jc ON jc.id = jca.category_id
WHERE jca.status = 'pending'
ORDER BY jca.submitted_at, jca.id
LIMIT sqlc.arg(p_limit);

-- name: ListDashboardExpiringRanks :many
SELECT
    urd.id,
    urd.expires_at,
    u.id AS user_id,
    u.last_name,
    u.first_name,
    u.middle_name,
    r.name AS credential_name,
    d.name AS discipline_name
FROM user_rank_discipline AS urd
JOIN "user" AS u ON u.id = urd.user_id
JOIN rank AS r ON r.id = urd.rank_id
JOIN discipline AS d ON d.id = urd.discipline_id
WHERE urd.expires_at BETWEEN CURRENT_DATE AND CURRENT_DATE + sqlc.arg(p_days)::integer
ORDER BY urd.expires_at, urd.id
LIMIT sqlc.arg(p_limit);

-- name: ListDashboardExpiringJudgeCategories :many
SELECT
    ujc.id,
    ujc.expires_at,
    u.id AS user_id,
    u.last_name,
    u.first_name,
    u.middle_name,
    jc.name AS credential_name
FROM user_judge_category AS ujc
JOIN "user" AS u ON u.id = ujc.user_id
JOIN judge_category AS jc ON jc.id = ujc.category_id
WHERE ujc.expires_at BETWEEN CURRENT_DATE AND CURRENT_DATE + sqlc.arg(p_days)::integer
ORDER BY ujc.expires_at, ujc.id
LIMIT sqlc.arg(p_limit);

-- name: ListDashboardRegistrationDeadlines :many
SELECT
    c.id AS competition_id,
    c.name AS competition_name,
    c.registration_ends_at AS due_at
FROM competition AS c
WHERE c.registration_ends_at BETWEEN now() AND now() + make_interval(days => sqlc.arg(p_days)::integer)
  AND c.status IN ('published', 'registration_open')
ORDER BY c.registration_ends_at, c.id
LIMIT sqlc.arg(p_limit);

-- name: ListDashboardReportDeadlines :many
SELECT id, name, due_at, status
FROM report_deadline
WHERE due_at >= now() AND status = 'planned'
ORDER BY due_at, id
LIMIT sqlc.arg(p_limit);
