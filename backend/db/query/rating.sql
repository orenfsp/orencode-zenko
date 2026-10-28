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
