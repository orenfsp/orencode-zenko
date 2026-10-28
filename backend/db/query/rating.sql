-- name: ListAthleteRating :many
WITH scored AS (
    SELECT
        u.id AS user_id,
        u.last_name,
        u.first_name,
        u.middle_name,
        m.name AS municipality_name,
        d.id AS discipline_id,
        d.name AS discipline_name,
        cr.id AS result_id,
        c.id AS competition_id,
        c.name AS competition_name,
        c.start_at,
        cr.place,
        cr.score,
        COALESCE(
            cr.score::double precision,
            CASE cr.place
                WHEN 1 THEN 100
                WHEN 2 THEN 80
                WHEN 3 THEN 60
                WHEN 4 THEN 50
                WHEN 5 THEN 40
                WHEN 6 THEN 30
                WHEN 7 THEN 20
                WHEN 8 THEN 10
                ELSE 0
            END
        ) AS points
    FROM competition_result AS cr
    JOIN competition_application AS ca ON ca.id = cr.application_id
    JOIN competition_event AS ce ON ce.id = ca.event_id
    JOIN competition AS c ON c.id = ce.competition_id
    JOIN "user" AS u ON u.id = ca.athlete_user_id
    JOIN discipline AS d ON d.id = ce.discipline_id
    LEFT JOIN municipality AS m ON m.id = u.municipality_id
    WHERE cr.result_status = 'finished'
      AND c.start_at >= make_date(sqlc.arg(year)::integer, 1, 1)
      AND c.start_at < make_date(sqlc.arg(year)::integer + 1, 1, 1)
      AND (sqlc.narg(discipline_id)::bigint IS NULL OR d.id = sqlc.narg(discipline_id)::bigint)
      AND (sqlc.narg(municipality_id)::bigint IS NULL OR u.municipality_id = sqlc.narg(municipality_id)::bigint)
      AND (sqlc.narg(age_group_id)::bigint IS NULL OR ce.age_group_id = sqlc.narg(age_group_id)::bigint)
), totals AS (
    SELECT user_id, last_name, first_name, middle_name, municipality_name,
           SUM(points)::double precision AS total_points,
           COUNT(*)::bigint AS result_count,
           COUNT(*) FILTER (WHERE place = 1)::bigint AS wins
    FROM scored
    GROUP BY user_id, last_name, first_name, middle_name, municipality_name
)
SELECT ROW_NUMBER() OVER (ORDER BY total_points DESC, wins DESC, result_count DESC, last_name, first_name, user_id)::bigint AS position,
       user_id, last_name, first_name, middle_name, municipality_name,
       total_points, result_count, wins
FROM totals
ORDER BY position
LIMIT sqlc.arg(p_limit)::integer OFFSET sqlc.arg(p_offset)::integer;

-- name: RatingDisciplineExists :one
SELECT EXISTS (SELECT 1 FROM discipline WHERE id = sqlc.arg(discipline_id));

-- name: RatingAgeGroupExists :one
SELECT EXISTS (SELECT 1 FROM age_group WHERE id = sqlc.arg(age_group_id));

-- name: RatingUserHasRole :one
SELECT EXISTS (
    SELECT 1
    FROM "user" AS u
    JOIN role_user AS ru ON ru.user_id = u.id
    JOIN role AS r ON r.id = ru.role_id
    WHERE u.id = sqlc.arg(user_id)
      AND u.access_status = true
      AND r.code = sqlc.arg(role_code)
);

-- name: UpsertRatingSelection :one
INSERT INTO rating_selection (
    season_year, discipline_id, age_group_id, captain_user_id, coach_user_id, created_by_user_id
) VALUES (
    sqlc.arg(season_year), sqlc.arg(discipline_id), sqlc.arg(age_group_id),
    sqlc.narg(captain_user_id), sqlc.narg(coach_user_id), sqlc.arg(created_by_user_id)
)
ON CONFLICT (season_year, discipline_id, age_group_id) DO UPDATE
SET captain_user_id = EXCLUDED.captain_user_id,
    coach_user_id = EXCLUDED.coach_user_id,
    updated_at = now()
WHERE rating_selection.status = 'draft'
RETURNING id;

-- name: DeleteRatingSelectionMembers :exec
DELETE FROM rating_selection_member WHERE selection_id = sqlc.arg(selection_id);

-- name: AddRatingSelectionMember :exec
INSERT INTO rating_selection_member (selection_id, user_id, selection_position)
VALUES (sqlc.arg(selection_id), sqlc.arg(user_id), sqlc.arg(selection_position));

-- name: GetRatingSelectionIDByScope :one
SELECT id
FROM rating_selection
WHERE season_year = sqlc.arg(season_year)
  AND discipline_id = sqlc.arg(discipline_id)
  AND age_group_id = sqlc.arg(age_group_id);

-- name: GetRatingSelection :one
SELECT
    rs.id, rs.season_year, rs.discipline_id, d.name AS discipline_name,
    rs.age_group_id, ag.name AS age_group_name, rs.status,
    rs.captain_user_id, rs.coach_user_id, rs.created_by_user_id,
    rs.locked_at, rs.created_at, rs.updated_at
FROM rating_selection AS rs
JOIN discipline AS d ON d.id = rs.discipline_id
JOIN age_group AS ag ON ag.id = rs.age_group_id
WHERE rs.id = sqlc.arg(selection_id);

-- name: ListRatingSelectionMembers :many
SELECT
    rsm.selection_position, u.id AS user_id,
    u.last_name, u.first_name, u.middle_name
FROM rating_selection_member AS rsm
JOIN "user" AS u ON u.id = rsm.user_id
WHERE rsm.selection_id = sqlc.arg(selection_id)
ORDER BY rsm.selection_position;

-- name: LockRatingSelection :one
UPDATE rating_selection
SET status = 'locked', locked_at = now(), updated_at = now()
WHERE id = sqlc.arg(selection_id) AND status = 'draft'
RETURNING id;

-- name: GetRatingSelectionStatus :one
SELECT status FROM rating_selection WHERE id = sqlc.arg(selection_id);
