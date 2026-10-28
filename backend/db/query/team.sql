-- name: CreateTeam :one
INSERT INTO team (name, team_type)
VALUES (sqlc.arg(name), 'regional_squad')
RETURNING id, name, team_type, created_at;

-- name: AddTeamMember :exec
INSERT INTO team_member (team_id, user_id, joined_at)
VALUES (sqlc.arg(team_id), sqlc.arg(user_id), CURRENT_DATE)
ON CONFLICT DO NOTHING;

-- name: AddTeamCoach :exec
INSERT INTO team_coach (team_id, coach_user_id)
VALUES (sqlc.arg(team_id), sqlc.arg(coach_user_id))
ON CONFLICT DO NOTHING;

-- name: ListTeams :many
SELECT id, name, team_type, created_at
FROM team
WHERE team_type = 'regional_squad'
ORDER BY created_at DESC, id DESC;

-- name: ListTeamMembers :many
SELECT tm.user_id, u.last_name, u.first_name, u.middle_name
FROM team_member AS tm
JOIN "user" AS u ON u.id = tm.user_id
WHERE tm.team_id = sqlc.arg(team_id)
  AND (tm.left_at IS NULL OR tm.left_at >= CURRENT_DATE)
ORDER BY u.last_name, u.first_name, tm.user_id;

-- name: ListTeamCoaches :many
SELECT tc.coach_user_id, u.last_name, u.first_name, u.middle_name
FROM team_coach AS tc
JOIN "user" AS u ON u.id = tc.coach_user_id
WHERE tc.team_id = sqlc.arg(team_id)
ORDER BY u.last_name, u.first_name, tc.coach_user_id;
