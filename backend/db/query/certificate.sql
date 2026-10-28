-- name: GetCertificateResult :one
SELECT cr.id AS result_id, ca.id AS application_id, ca.athlete_user_id,
       ca.status AS application_status, ce.id AS event_id, ce.discipline_id,
       c.id AS competition_id, c.name AS competition_name, c.start_at, c.end_at,
       d.name AS discipline_name, u.first_name, u.last_name, u.middle_name,
       org.name AS organization_name, cr.place, cr.result_status,
       sheet.id AS sheet_id, sheet.mode, sheet.revision, sheet.approved_at,
       entry.id AS entry_id, entry.place AS entry_place, entry.result_status AS entry_status,
       team.name AS team_name
FROM competition_result cr
JOIN competition_application ca ON ca.id = cr.application_id
JOIN competition_event ce ON ce.id = ca.event_id
JOIN competition c ON c.id = ce.competition_id
JOIN discipline d ON d.id = ce.discipline_id
JOIN "user" u ON u.id = ca.athlete_user_id
LEFT JOIN organization org ON org.id = u.organization_id
JOIN competition_event_result_entry_application link ON link.application_id = ca.id
JOIN competition_event_result_entry entry ON entry.id = link.entry_id
JOIN competition_event_result_sheet sheet ON sheet.id = entry.sheet_id AND sheet.event_id = ce.id
LEFT JOIN team ON team.id = entry.team_id
WHERE cr.id = sqlc.arg(result_id)
  AND sheet.status = 'approved' AND sheet.approved_at IS NOT NULL
  AND ca.status = 'approved' AND cr.result_status IS NOT NULL
  AND cr.result_status = entry.result_status
  AND cr.place IS NOT DISTINCT FROM entry.place;

-- name: CertificateConfirmedCoach :one
SELECT EXISTS (SELECT 1 FROM user_coach
 WHERE user_id = sqlc.arg(athlete_id) AND coach_user_id = sqlc.arg(coach_id)
 AND discipline_id = sqlc.arg(discipline_id) AND status = 'confirmed' AND ended_at IS NULL);
