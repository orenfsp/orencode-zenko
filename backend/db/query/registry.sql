-- name: ListRegistryUsers :many
SELECT
    u.id,
    u.last_name,
    u.first_name,
    u.middle_name,
    u.birth,
    o.id AS organization_id,
    o.name AS organization_name,
    m.id AS municipality_id,
    m.name AS municipality_name,
    COALESCE((
        SELECT jsonb_agg(jsonb_build_object('id', r.id, 'code', r.code, 'name', r.name) ORDER BY r.name, r.id)
        FROM role_user AS ru
        JOIN role AS r ON r.id = ru.role_id
        WHERE ru.user_id = u.id
    ), '[]'::jsonb) AS roles,
    COALESCE((
        SELECT jsonb_agg(jsonb_build_object('id', d.id, 'name', d.name) ORDER BY d.name, d.id)
        FROM user_discipline AS ud
        JOIN discipline AS d ON d.id = ud.discipline_id
        WHERE ud.user_id = u.id
    ), '[]'::jsonb) AS disciplines,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', urd.id,
                'rank_id', r.id,
                'rank_name', r.name,
                'discipline_id', d.id,
                'discipline_name', d.name,
                'awarded_at', urd.awarded_at,
                'expires_at', urd.expires_at,
                'order_number', urd.order_number
            ) ORDER BY urd.awarded_at DESC NULLS LAST, urd.id DESC
        )
        FROM user_rank_discipline AS urd
        JOIN rank AS r ON r.id = urd.rank_id
        JOIN discipline AS d ON d.id = urd.discipline_id
        WHERE urd.user_id = u.id
          AND (urd.expires_at IS NULL OR urd.expires_at >= CURRENT_DATE)
    ), '[]'::jsonb) AS ranks,
    COALESCE((
        SELECT jsonb_build_object(
            'id', ujc.id,
            'category_id', jc.id,
            'category_name', jc.name,
            'assigned_at', ujc.assigned_at,
            'expires_at', ujc.expires_at,
            'order_number', ujc.order_number
        )
        FROM user_judge_category AS ujc
        JOIN judge_category AS jc ON jc.id = ujc.category_id
        WHERE ujc.user_id = u.id
          AND (ujc.expires_at IS NULL OR ujc.expires_at >= CURRENT_DATE)
        ORDER BY ujc.assigned_at DESC, ujc.id DESC
        LIMIT 1
    ), 'null'::jsonb) AS judge_category,
    COALESCE((
        SELECT jsonb_agg(rel.item ORDER BY rel.requested_at DESC, rel.id DESC)
        FROM (
            SELECT
                uc.id,
                uc.requested_at,
                jsonb_build_object(
                    'id', uc.id,
                    'relation_type', 'participant',
                    'related_user_id', participant.id,
                    'related_user_name', concat_ws(' ', participant.last_name, participant.first_name, participant.middle_name),
                    'discipline', jsonb_build_object('id', d.id, 'name', d.name),
                    'status', uc.status,
                    'requested_at', uc.requested_at,
                    'confirmed_at', uc.confirmed_at,
                    'ended_at', uc.ended_at
                ) AS item
            FROM user_coach AS uc
            JOIN "user" AS participant ON participant.id = uc.user_id
            JOIN discipline AS d ON d.id = uc.discipline_id
            WHERE uc.coach_user_id = u.id
            UNION ALL
            SELECT
                uc.id,
                uc.requested_at,
                jsonb_build_object(
                    'id', uc.id,
                    'relation_type', 'coach',
                    'related_user_id', coach.id,
                    'related_user_name', concat_ws(' ', coach.last_name, coach.first_name, coach.middle_name),
                    'discipline', jsonb_build_object('id', d.id, 'name', d.name),
                    'status', uc.status,
                    'requested_at', uc.requested_at,
                    'confirmed_at', uc.confirmed_at,
                    'ended_at', uc.ended_at
                ) AS item
            FROM user_coach AS uc
            JOIN "user" AS coach ON coach.id = uc.coach_user_id
            JOIN discipline AS d ON d.id = uc.discipline_id
            WHERE uc.user_id = u.id
        ) AS rel
    ), '[]'::jsonb) AS coach_relations,
    COALESCE((
        SELECT jsonb_agg(
            jsonb_build_object(
                'id', cj.id,
                'competition_id', c.id,
                'competition_name', c.name,
                'competition_status', c.status,
                'position_id', jp.id,
                'position_name', jp.name,
                'start_at', c.start_at,
                'end_at', c.end_at,
                'appointed_at', cj.appointed_at
            ) ORDER BY c.start_at DESC, cj.id DESC
        )
        FROM competition_judge AS cj
        JOIN competition AS c ON c.id = cj.competition_id
        JOIN judge_position AS jp ON jp.id = cj.position_id
        WHERE cj.judge_user_id = u.id
    ), '[]'::jsonb) AS judge_assignments,
    COUNT(*) OVER() AS total_count
FROM "user" AS u
LEFT JOIN organization AS o ON o.id = u.organization_id
LEFT JOIN municipality AS m ON m.id = u.municipality_id
WHERE u.access_status = true
  AND (
      sqlc.narg(p_search)::text IS NULL
      OR regexp_replace(concat_ws(' ', u.last_name, u.first_name, u.middle_name), '[[:space:]]+', ' ', 'g') ILIKE '%' || sqlc.narg(p_search) || '%'
  )
  AND (sqlc.narg(p_organization_id)::bigint IS NULL OR u.organization_id = sqlc.narg(p_organization_id))
  AND (sqlc.narg(p_municipality_id)::bigint IS NULL OR u.municipality_id = sqlc.narg(p_municipality_id))
  AND (
      sqlc.narg(p_role_code)::text IS NULL
      OR EXISTS (
          SELECT 1 FROM role_user AS ru
          JOIN role AS r ON r.id = ru.role_id
          WHERE ru.user_id = u.id AND r.code = sqlc.narg(p_role_code)
      )
  )
  AND (
      sqlc.narg(p_discipline_id)::bigint IS NULL
      OR EXISTS (
          SELECT 1 FROM user_discipline AS ud
          WHERE ud.user_id = u.id AND ud.discipline_id = sqlc.narg(p_discipline_id)
      )
  )
  AND (
      sqlc.narg(p_rank_id)::bigint IS NULL
      OR EXISTS (
          SELECT 1 FROM user_rank_discipline AS urd
          WHERE urd.user_id = u.id
            AND urd.rank_id = sqlc.narg(p_rank_id)
            AND (urd.expires_at IS NULL OR urd.expires_at >= CURRENT_DATE)
      )
  )
  AND (
      sqlc.narg(p_age_group_id)::bigint IS NULL
      OR EXISTS (
          SELECT 1
          FROM age_group AS ag
          WHERE ag.id = sqlc.narg(p_age_group_id)
            AND (ag.min_age IS NULL OR date_part('year', age(CURRENT_DATE, u.birth)) >= ag.min_age)
            AND (ag.max_age IS NULL OR date_part('year', age(CURRENT_DATE, u.birth)) <= ag.max_age)
      )
  )
ORDER BY u.last_name, u.first_name, u.middle_name NULLS FIRST, u.id
LIMIT sqlc.arg(p_limit) OFFSET sqlc.arg(p_offset);

-- name: ListRegistryOrganizations :many
SELECT
    o.id,
    o.name,
    o.organization_type,
    m.id AS municipality_id,
    m.name AS municipality_name,
    COUNT(*) OVER() AS total_count
FROM organization AS o
LEFT JOIN municipality AS m ON m.id = o.municipality_id
WHERE (
    sqlc.narg(p_search)::text IS NULL
    OR o.name ILIKE '%' || sqlc.narg(p_search) || '%'
)
AND (sqlc.narg(p_municipality_id)::bigint IS NULL OR o.municipality_id = sqlc.narg(p_municipality_id))
ORDER BY o.name, o.id
LIMIT sqlc.arg(p_limit) OFFSET sqlc.arg(p_offset);
