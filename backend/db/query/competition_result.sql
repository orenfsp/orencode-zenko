-- name: GetResultEventContext :one
SELECT
    ce.id AS event_id,
    c.id AS competition_id,
    c.name AS competition_name,
    c.status AS competition_status,
    c.start_at,
    c.end_at,
    d.id AS discipline_id,
    d.name AS discipline_name,
    ag.name AS age_group_name
FROM competition_event AS ce
JOIN competition AS c ON c.id = ce.competition_id
JOIN discipline AS d ON d.id = ce.discipline_id
LEFT JOIN age_group AS ag ON ag.id = ce.age_group_id
WHERE ce.id = sqlc.arg(p_event_id)
  AND ce.competition_id = sqlc.arg(p_competition_id);

-- name: ListApprovedResultApplications :many
SELECT
    ca.id AS application_id,
    ca.athlete_user_id,
    u.last_name,
    u.first_name,
    u.middle_name
FROM competition_application AS ca
JOIN "user" AS u ON u.id = ca.athlete_user_id
WHERE ca.event_id = sqlc.arg(p_event_id)
  AND ca.status = 'approved'
ORDER BY u.last_name, u.first_name, u.middle_name, ca.id;

-- name: ListEligibleResultTeamMemberships :many
SELECT
    ca.id AS application_id,
    ca.athlete_user_id,
    t.id AS team_id,
    t.name AS team_name
FROM competition_application AS ca
JOIN competition_event AS ce ON ce.id = ca.event_id
JOIN competition AS c ON c.id = ce.competition_id
JOIN team_member AS tm
  ON tm.user_id = ca.athlete_user_id
 AND (tm.joined_at IS NULL OR tm.joined_at <= c.start_at::date)
 AND (tm.left_at IS NULL OR tm.left_at >= c.start_at::date)
JOIN team AS t ON t.id = tm.team_id
WHERE ca.event_id = sqlc.arg(p_event_id)
  AND ca.status = 'approved'
ORDER BY t.name, t.id, ca.id;

-- name: GetEventResultSheet :one
SELECT *
FROM competition_event_result_sheet
WHERE event_id = sqlc.arg(p_event_id);

-- name: LockEventResultSheet :one
SELECT *
FROM competition_event_result_sheet
WHERE event_id = sqlc.arg(p_event_id)
FOR UPDATE;

-- name: CreateEventResultSheet :one
INSERT INTO competition_event_result_sheet (event_id, mode, task_count)
VALUES (sqlc.arg(p_event_id), sqlc.arg(p_mode), sqlc.arg(p_task_count))
RETURNING *;

-- name: UpdateEventResultSheetSettings :one
UPDATE competition_event_result_sheet
SET mode = sqlc.arg(p_mode),
    task_count = sqlc.arg(p_task_count),
    scoring_mode = 'algorithmic',
    template_id = NULL,
    template_snapshot = NULL,
    calculated_at = NULL,
    revision = revision + 1,
    updated_at = now()
WHERE event_id = sqlc.arg(p_event_id)
  AND status = 'draft'
RETURNING *;

-- name: DeleteEventResultEntries :exec
DELETE FROM competition_event_result_entry
WHERE sheet_id = sqlc.arg(p_sheet_id);

-- name: CreateEventResultEntry :one
INSERT INTO competition_event_result_entry (sheet_id, team_id)
VALUES (sqlc.arg(p_sheet_id), sqlc.narg(p_team_id))
RETURNING id;

-- name: LinkResultEntryApplication :exec
INSERT INTO competition_event_result_entry_application (entry_id, application_id)
VALUES (sqlc.arg(p_entry_id), sqlc.arg(p_application_id));

-- name: ListEventResultEntries :many
SELECT
    e.id,
    e.team_id,
    t.name AS team_name,
    e.result_status,
    e.solved_count,
    e.total_time_deciseconds,
    e.last_solve_time_deciseconds,
    e.negative_attempts,
    e.place
    , e.custom_values
FROM competition_event_result_entry AS e
LEFT JOIN team AS t ON t.id = e.team_id
WHERE e.sheet_id = sqlc.arg(p_sheet_id)
ORDER BY e.id;

-- name: ListResultEntryApplications :many
SELECT
    link.entry_id,
    ca.id AS application_id,
    ca.athlete_user_id,
    u.last_name,
    u.first_name,
    u.middle_name
FROM competition_event_result_entry_application AS link
JOIN competition_application AS ca ON ca.id = link.application_id
JOIN "user" AS u ON u.id = ca.athlete_user_id
JOIN competition_event_result_entry AS e ON e.id = link.entry_id
WHERE e.sheet_id = sqlc.arg(p_sheet_id)
ORDER BY link.entry_id, u.last_name, u.first_name, u.middle_name, ca.id;

-- name: ListEventTaskResults :many
SELECT
    tr.entry_id,
    tr.task_number,
    tr.status,
    tr.solve_time_deciseconds,
    tr.negative_attempts
FROM competition_event_task_result AS tr
JOIN competition_event_result_entry AS e ON e.id = tr.entry_id
WHERE e.sheet_id = sqlc.arg(p_sheet_id)
ORDER BY tr.entry_id, tr.task_number;

-- name: SetResultEntryStatus :execrows
UPDATE competition_event_result_entry AS e
SET result_status = sqlc.arg(p_result_status),
    solved_count = 0,
    total_time_deciseconds = 0,
    last_solve_time_deciseconds = 0,
    negative_attempts = 0,
    place = NULL
FROM competition_event_result_sheet AS s
WHERE e.id = sqlc.arg(p_entry_id)
  AND e.sheet_id = s.id
  AND s.id = sqlc.arg(p_sheet_id)
  AND s.status = 'draft';

-- name: DeleteResultEntryTasks :exec
DELETE FROM competition_event_task_result AS tr
USING competition_event_result_entry AS e
WHERE tr.entry_id = e.id
  AND e.id = sqlc.arg(p_entry_id)
  AND e.sheet_id = sqlc.arg(p_sheet_id);

-- name: UpsertEventTaskResult :exec
INSERT INTO competition_event_task_result (
    entry_id,
    task_number,
    status,
    solve_time_deciseconds,
    negative_attempts
)
VALUES (
    sqlc.arg(p_entry_id),
    sqlc.arg(p_task_number),
    sqlc.arg(p_status),
    sqlc.narg(p_solve_time_deciseconds),
    sqlc.arg(p_negative_attempts)
)
ON CONFLICT (entry_id, task_number) DO UPDATE
SET status = EXCLUDED.status,
    solve_time_deciseconds = EXCLUDED.solve_time_deciseconds,
    negative_attempts = EXCLUDED.negative_attempts;

-- name: UpdateResultEntrySummary :execrows
UPDATE competition_event_result_entry AS e
SET solved_count = sqlc.arg(p_solved_count),
    total_time_deciseconds = sqlc.arg(p_total_time_deciseconds),
    last_solve_time_deciseconds = sqlc.arg(p_last_solve_time_deciseconds),
    negative_attempts = sqlc.arg(p_negative_attempts),
    place = sqlc.arg(p_place)
FROM competition_event_result_sheet AS s
WHERE e.id = sqlc.arg(p_entry_id)
  AND e.sheet_id = s.id
  AND s.id = sqlc.arg(p_sheet_id)
  AND s.status = 'draft';

-- name: MarkEventResultSheetCalculated :execrows
UPDATE competition_event_result_sheet
SET calculated_at = now(),
    revision = revision + 1,
    updated_at = now()
WHERE id = sqlc.arg(p_sheet_id)
  AND status = 'draft';

-- name: MarkEventResultSheetDirty :exec
UPDATE competition_event_result_sheet
SET calculated_at = NULL,
    revision = revision + 1,
    updated_at = now()
WHERE id = sqlc.arg(p_sheet_id)
  AND status = 'draft';

-- name: ClearEventResultSummaries :exec
UPDATE competition_event_result_entry
SET solved_count = 0, total_time_deciseconds = 0,
    last_solve_time_deciseconds = 0, negative_attempts = 0, place = NULL
WHERE sheet_id = sqlc.arg(p_sheet_id);

-- name: UpsertApprovedCompetitionResult :one
INSERT INTO competition_result (application_id, place, score, result_status)
VALUES (
    sqlc.arg(p_application_id),
    sqlc.arg(p_place),
    sqlc.arg(p_score),
    sqlc.arg(p_result_status)
)
ON CONFLICT (application_id) DO UPDATE
SET place = EXCLUDED.place,
    score = EXCLUDED.score,
    result_status = EXCLUDED.result_status
RETURNING id;

-- name: ApproveEventResultSheet :one
UPDATE competition_event_result_sheet
SET status = 'approved',
    revision = revision + 1,
    calculated_at = now(),
    approved_by_user_id = sqlc.arg(p_approved_by_user_id),
    approved_at = now(),
    protocol_storage_key = sqlc.arg(p_protocol_storage_key),
    updated_at = now()
WHERE id = sqlc.arg(p_sheet_id)
  AND status = 'draft'
RETURNING approved_at;

-- name: SetCompetitionEventProtocolURL :exec
UPDATE competition_event
SET protocol_url = sqlc.arg(p_protocol_url)
WHERE id = sqlc.arg(p_event_id);

-- name: SetResultProtocolStorageKey :exec
UPDATE competition_event_result_sheet
SET protocol_storage_key = sqlc.narg(p_protocol_storage_key), updated_at = now()
WHERE id = sqlc.arg(p_sheet_id);

-- name: ListResultProtocolJudges :many
SELECT
    lower(jp.name) AS position_name,
    u.last_name,
    u.first_name,
    u.middle_name
FROM competition_judge AS cj
JOIN judge_position AS jp ON jp.id = cj.position_id
JOIN "user" AS u ON u.id = cj.judge_user_id
WHERE cj.competition_id = sqlc.arg(p_competition_id)
  AND lower(jp.name) IN ('главный судья', 'главный секретарь')
ORDER BY lower(jp.name), u.last_name, u.first_name, u.id;

-- name: ListAthleteResultHistory :many
SELECT cr.id AS result_id, c.id AS competition_id, c.name AS competition_name, ce.id AS event_id,
       d.id AS discipline_id, d.name AS discipline_name, c.start_at,
       cr.place, cr.score, cr.result_status, ce.protocol_url
FROM competition_result cr
JOIN competition_application ca ON ca.id = cr.application_id
JOIN competition_event ce ON ce.id = ca.event_id
JOIN competition c ON c.id = ce.competition_id
JOIN discipline d ON d.id = ce.discipline_id
WHERE ca.athlete_user_id = sqlc.arg(p_user_id)
  AND (sqlc.narg(p_discipline_id)::bigint IS NULL OR ce.discipline_id = sqlc.narg(p_discipline_id))
ORDER BY c.start_at DESC, ce.id DESC, cr.id DESC;
