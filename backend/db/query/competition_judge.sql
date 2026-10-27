-- name: ListCompetitionJudges :many
SELECT
    cj.id,
    cj.competition_id,
    cj.judge_user_id,
    u.first_name,
    u.last_name,
    u.middle_name,
    cj.position_id,
    jp.name AS position_name,
    cj.appointed_at
FROM competition_judge AS cj
JOIN competition AS c
    ON c.id = cj.competition_id
JOIN "user" AS u
    ON u.id = cj.judge_user_id
JOIN judge_position AS jp
    ON jp.id = cj.position_id
WHERE cj.competition_id = sqlc.arg(p_competition_id)
  AND c.status = ANY(sqlc.arg(p_public_statuses)::text[])
ORDER BY jp.name, u.last_name, u.first_name, cj.id;


-- name: AssignCompetitionJudge :one
WITH target_competition AS (
    SELECT c.id
    FROM competition AS c
    WHERE c.id = sqlc.arg(p_competition_id)
      AND c.status = 'draft'
    FOR UPDATE
)
INSERT INTO competition_judge (
    competition_id,
    judge_user_id,
    position_id
)
SELECT
    target.id,
    sqlc.arg(p_judge_user_id),
    sqlc.arg(p_position_id)
FROM target_competition AS target
WHERE NOT EXISTS (
    SELECT 1
    FROM competition_judge AS existing
    WHERE existing.competition_id = target.id
      AND existing.judge_user_id = sqlc.arg(p_judge_user_id)
      AND existing.position_id = sqlc.arg(p_position_id)
)
RETURNING
    id,
    competition_id,
    judge_user_id,
    position_id,
    appointed_at;


-- name: RemoveCompetitionJudge :execrows
WITH target_competition AS (
    SELECT c.id
    FROM competition AS c
    WHERE c.id = sqlc.arg(p_competition_id)
      AND c.status = 'draft'
    FOR UPDATE
)
DELETE FROM competition_judge AS cj
USING target_competition AS target
WHERE cj.competition_id = target.id
  AND cj.id = sqlc.arg(p_assignment_id);

-- name: ValidateCompetitionJudgeAssignment :one
SELECT
    EXISTS (
        SELECT 1 FROM competition AS c WHERE c.id = sqlc.arg(p_competition_id)
    ) AS competition_exists,
    EXISTS (
        SELECT 1 FROM "user" AS u WHERE u.id = sqlc.arg(p_judge_user_id)
    ) AS user_exists,
    EXISTS (
        SELECT 1
        FROM role_user AS ru
        JOIN role AS r ON r.id = ru.role_id
        WHERE ru.user_id = sqlc.arg(p_judge_user_id)
          AND r.code = sqlc.arg(p_judge_role_code)
    ) AS judge_role_exists,
    EXISTS (
        SELECT 1 FROM judge_position AS jp WHERE jp.id = sqlc.arg(p_position_id)
    ) AS position_exists,
    EXISTS (
        SELECT 1
        FROM competition_judge AS cj
        WHERE cj.competition_id = sqlc.arg(p_competition_id)
          AND cj.judge_user_id = sqlc.arg(p_judge_user_id)
          AND cj.position_id = sqlc.arg(p_position_id)
    ) AS assignment_exists;
