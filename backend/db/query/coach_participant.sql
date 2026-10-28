-- name: SearchCoachParticipants :many
SELECT u.id, u.first_name, u.last_name, u.middle_name,
       o.name AS organization, m.name AS municipality,
       COALESCE((SELECT jsonb_agg(jsonb_build_object('id', d.id, 'name', d.name) ORDER BY d.name, d.id)
                 FROM user_discipline ud JOIN discipline d ON d.id = ud.discipline_id
                 WHERE ud.user_id = u.id), '[]'::jsonb) AS disciplines,
       COALESCE((SELECT jsonb_agg(jsonb_build_object('discipline_id', relation.discipline_id, 'status', relation.status)
                                  ORDER BY relation.discipline_id)
                 FROM (SELECT DISTINCT ON (uc.discipline_id) uc.discipline_id, uc.status
                       FROM user_coach uc
                       WHERE uc.user_id = u.id AND uc.coach_user_id = sqlc.arg(coach_id)
                       ORDER BY uc.discipline_id, (uc.status IN ('pending', 'confirmed')) DESC,
                                uc.requested_at DESC, uc.id DESC) relation), '[]'::jsonb) AS coach_relations
FROM "user" u
LEFT JOIN organization o ON o.id = u.organization_id
LEFT JOIN municipality m ON m.id = u.municipality_id
WHERE u.access_status AND u.id <> sqlc.arg(coach_id)
  AND EXISTS (SELECT 1 FROM role_user ru JOIN role r ON r.id = ru.role_id WHERE ru.user_id = u.id AND r.code = 'participant')
  AND (sqlc.arg(search)::text = '' OR concat_ws(' ', u.last_name, u.first_name, u.middle_name) ILIKE '%' || sqlc.arg(search)::text || '%'
       OR concat_ws(' ', u.first_name, u.last_name, u.middle_name) ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (sqlc.narg(discipline_id)::bigint IS NULL OR EXISTS (SELECT 1 FROM user_discipline ud WHERE ud.user_id = u.id AND ud.discipline_id = sqlc.narg(discipline_id)))
ORDER BY u.last_name, u.first_name, u.id
LIMIT sqlc.arg(page_limit)::integer OFFSET sqlc.arg(page_offset)::integer;

-- name: CountCoachParticipants :one
SELECT count(*) FROM "user" u
WHERE u.access_status AND u.id <> sqlc.arg(coach_id)
  AND EXISTS (SELECT 1 FROM role_user ru JOIN role r ON r.id = ru.role_id WHERE ru.user_id = u.id AND r.code = 'participant')
  AND (sqlc.arg(search)::text = '' OR concat_ws(' ', u.last_name, u.first_name, u.middle_name) ILIKE '%' || sqlc.arg(search)::text || '%'
       OR concat_ws(' ', u.first_name, u.last_name, u.middle_name) ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (sqlc.narg(discipline_id)::bigint IS NULL OR EXISTS (SELECT 1 FROM user_discipline ud WHERE ud.user_id = u.id AND ud.discipline_id = sqlc.narg(discipline_id)));
