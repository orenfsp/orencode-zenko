package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/orenfsp/orencode-zenko.git/config"
	postgres "github.com/orenfsp/orencode-zenko.git/pkg/db"
	"golang.org/x/crypto/bcrypt"
)

const (
	demoDisciplineName = "программирование алгоритмическое"
	demoPassword       = "Password123!"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config load:", err)
	}

	ctx := context.Background()
	pool := postgres.NewPostgres(ctx, cfg.Database.URL)
	defer pool.Close()

	mustExec := func(label, query string, args ...any) {
		if _, execErr := pool.Exec(ctx, query, args...); execErr != nil {
			log.Fatalf("%s: %v", label, execErr)
		}
	}

	getOrCreateNamedID := func(table, name string) int64 {
		var id int64
		query := `SELECT id FROM ` + table + ` WHERE name = $1 ORDER BY id LIMIT 1`
		err = pool.QueryRow(ctx, query, name).Scan(&id)
		if err == nil {
			return id
		}
		if err != pgx.ErrNoRows {
			log.Fatalf("find %s: %v", table, err)
		}

		query = `INSERT INTO ` + table + ` (name) VALUES ($1) RETURNING id`
		if err = pool.QueryRow(ctx, query, name).Scan(&id); err != nil {
			log.Fatalf("create %s: %v", table, err)
		}
		return id
	}

	// Reference data used by the demo scenarios.
	disciplineID := getOrCreateNamedID("discipline", demoDisciplineName)
	typeID := getOrCreateNamedID("competition_type", "Чемпионат")
	levelID := getOrCreateNamedID("competition_level", "Региональный")
	municipalityID := getOrCreateNamedID("municipality", "г. Оренбург")
	rankID := getOrCreateNamedID("rank", "Кандидат в мастера спорта")
	judgeCategoryID := getOrCreateNamedID("judge_category", "Спортивный судья третьей категории")
	judgePositionID := getOrCreateNamedID("judge_position", "Главный судья")
	ageGroupID := getOrCreateNamedID("age_group", "Юниоры 16–25 лет")

	mustExec("update rank", `UPDATE rank SET ttl = 3 WHERE id = $1`, rankID)
	mustExec("update age group", `UPDATE age_group SET min_age = 16, max_age = 25 WHERE id = $1`, ageGroupID)

	var organizationID int64
	err = pool.QueryRow(ctx, `
		SELECT id FROM organization
		WHERE name = 'Оренбургский государственный университет'
		ORDER BY id LIMIT 1
	`).Scan(&organizationID)
	if err == pgx.ErrNoRows {
		err = pool.QueryRow(ctx, `
			INSERT INTO organization (name, organization_type, municipality_id)
			VALUES ('Оренбургский государственный университет', 'вуз', $1)
			RETURNING id
		`, municipalityID).Scan(&organizationID)
	} else if err == nil {
		mustExec("update organization", `
			UPDATE organization
			SET organization_type = 'вуз', municipality_id = $2
			WHERE id = $1
		`, organizationID, municipalityID)
	}
	if err != nil {
		log.Fatal("upsert organization:", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(demoPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("generate password hash:", err)
	}
	passwordHash := string(hash)

	createUser := func(first, last, birth, email string) int64 {
		var id int64
		if scanErr := pool.QueryRow(ctx, `
			INSERT INTO "user" (
				first_name, last_name, birth, email, password_hash,
				municipality_id, organization_id
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (email) DO UPDATE SET
				first_name = EXCLUDED.first_name,
				last_name = EXCLUDED.last_name,
				birth = EXCLUDED.birth,
				password_hash = EXCLUDED.password_hash,
				municipality_id = EXCLUDED.municipality_id,
				organization_id = EXCLUDED.organization_id,
				access_status = true
			RETURNING id
		`, first, last, birth, email, passwordHash, municipalityID, organizationID).Scan(&id); scanErr != nil {
			log.Fatalf("create user %s: %v", email, scanErr)
		}
		return id
	}

	athleteID := createUser("Иван", "Спортсменов", "2005-05-15", "athlete@example.local")
	athleteTwoID := createUser("Анна", "Быстрая", "2004-09-21", "athlete2@example.local")
	athleteThreeID := createUser("Максим", "Новичков", "2007-03-12", "athlete3@example.local")
	coachID := createUser("Петр", "Тренеров", "1980-01-01", "coach@example.local")
	judgeID := createUser("Сидор", "Судейкин", "1975-02-02", "judge@example.local")
	judgeCandidateID := createUser("Елена", "Арбитрова", "1988-07-08", "judge2@example.local")

	var adminID int64
	if err = pool.QueryRow(ctx, `
		SELECT u.id
		FROM "user" AS u
		JOIN role_user AS ru ON ru.user_id = u.id
		JOIN "role" AS r ON r.id = ru.role_id
		WHERE r.code = 'super_admin'
		ORDER BY u.id LIMIT 1
	`).Scan(&adminID); err != nil {
		log.Fatalf("find RBAC super administrator: %v", err)
	}

	assignRole := func(userID int64, roleCode string) {
		mustExec("assign role "+roleCode, `
			INSERT INTO role_user (user_id, role_id)
			SELECT $1, id FROM "role" WHERE code = $2
			ON CONFLICT (user_id, role_id) DO NOTHING
		`, userID, roleCode)
	}
	for _, userID := range []int64{athleteID, athleteTwoID, athleteThreeID} {
		assignRole(userID, "participant")
	}
	assignRole(coachID, "coach")
	assignRole(judgeID, "judge")
	assignRole(judgeCandidateID, "judge")

	for _, userID := range []int64{athleteID, athleteTwoID, athleteThreeID, coachID, judgeID, judgeCandidateID} {
		mustExec("assign discipline", `
			INSERT INTO user_discipline (user_id, discipline_id, start_year)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, discipline_id) DO UPDATE
			SET start_year = EXCLUDED.start_year
		`, userID, disciplineID, 2022)
	}

	upsertCoachRelation := func(athleteUserID int64, status string) {
		var relationID int64
		scanErr := pool.QueryRow(ctx, `
			SELECT id FROM user_coach
			WHERE user_id = $1 AND coach_user_id = $2 AND discipline_id = $3
			  AND status IN ('pending', 'confirmed')
			ORDER BY id DESC LIMIT 1
		`, athleteUserID, coachID, disciplineID).Scan(&relationID)
		if scanErr == pgx.ErrNoRows {
			mustExec("create coach relation", `
				INSERT INTO user_coach (
					user_id, coach_user_id, discipline_id, status, confirmed_at
				)
				VALUES (
					$1, $2, $3, $4::varchar,
					CASE WHEN $4::varchar = 'confirmed' THEN now() END
				)
			`, athleteUserID, coachID, disciplineID, status)
			return
		}
		if scanErr != nil {
			log.Fatalf("find coach relation: %v", scanErr)
		}
		mustExec("update coach relation", `
			UPDATE user_coach
			SET status = $2::varchar,
				confirmed_at = CASE WHEN $2::varchar = 'confirmed' THEN COALESCE(confirmed_at, now()) END,
				ended_at = NULL
			WHERE id = $1
		`, relationID, status)
	}
	upsertCoachRelation(athleteID, "confirmed")
	upsertCoachRelation(athleteTwoID, "confirmed")
	upsertCoachRelation(athleteThreeID, "pending")

	// Judge profile and a second pending application make both profile and review screens non-empty.
	mustExec("assign judge category", `
		INSERT INTO user_judge_category (
			user_id, category_id, assigned_at, expires_at, order_number
		)
		SELECT $1, $2, CURRENT_DATE - 180, CURRENT_DATE + 900, 'DEMO-JUDGE-001'
		WHERE NOT EXISTS (
			SELECT 1 FROM user_judge_category WHERE user_id = $1 AND category_id = $2
		)
	`, judgeID, judgeCategoryID)
	mustExec("create judge category application", `
		INSERT INTO judge_category_application (
			judge_user_id, category_id, application_type, status
		)
		SELECT $1, $2, 'assignment', 'pending'
		WHERE NOT EXISTS (
			SELECT 1 FROM judge_category_application
			WHERE judge_user_id = $1 AND category_id = $2
			  AND application_type = 'assignment' AND status = 'pending'
		)
	`, judgeCandidateID, judgeCategoryID)

	now := time.Now().UTC()
	year := int16(now.Year())
	var calendarPlanID int64
	err = pool.QueryRow(ctx, `
		SELECT id FROM calendar_plan
		WHERE name = 'Демонстрационный календарный план' AND year = $1
		ORDER BY id LIMIT 1
	`, year).Scan(&calendarPlanID)
	if err == pgx.ErrNoRows {
		err = pool.QueryRow(ctx, `
			INSERT INTO calendar_plan (name, year)
			VALUES ('Демонстрационный календарный план', $1)
			RETURNING id
		`, year).Scan(&calendarPlanID)
	}
	if err != nil {
		log.Fatal("upsert calendar plan:", err)
	}

	upsertCompetition := func(name, description, status string, startAt, endAt, registrationStartsAt, registrationEndsAt time.Time) int64 {
		var id int64
		if scanErr := pool.QueryRow(ctx, `
			INSERT INTO competition (
				name, description, discipline_id, competition_type_id,
				competition_level_id, format, start_at, end_at,
				registration_starts_at, registration_ends_at, online_url,
				created_by, status, calendar_plan_id, calendar_plan_number
			)
			VALUES (
				$1, $2, $3, $4, $5, 'online', $6, $7, $8, $9,
				'https://example.local/contest', $10, $11, $12, $13
			)
			ON CONFLICT (name) DO UPDATE SET
				description = EXCLUDED.description,
				discipline_id = EXCLUDED.discipline_id,
				competition_type_id = EXCLUDED.competition_type_id,
				competition_level_id = EXCLUDED.competition_level_id,
				format = EXCLUDED.format,
				start_at = EXCLUDED.start_at,
				end_at = EXCLUDED.end_at,
				registration_starts_at = EXCLUDED.registration_starts_at,
				registration_ends_at = EXCLUDED.registration_ends_at,
				status = EXCLUDED.status,
				calendar_plan_id = EXCLUDED.calendar_plan_id,
				calendar_plan_number = EXCLUDED.calendar_plan_number,
				updated_at = now()
			RETURNING id
		`, name, description, disciplineID, typeID, levelID,
			startAt, endAt, registrationStartsAt, registrationEndsAt,
			adminID, status, calendarPlanID, fmt.Sprintf("DEMO-%d", now.Year()),
		).Scan(&id); scanErr != nil {
			log.Fatalf("upsert competition %s: %v", name, scanErr)
		}
		return id
	}

	completedCompetitionID := upsertCompetition(
		"Кубок Оренбургской области",
		"Завершённое соревнование с утверждёнными результатами.",
		"completed",
		now.AddDate(0, -1, 0), now.AddDate(0, -1, 1),
		now.AddDate(0, -2, 0), now.AddDate(0, -1, -7),
	)
	upcomingCompetitionID := upsertCompetition(
		"Открытый чемпионат Оренбуржья",
		"Предстоящее соревнование с открытой регистрацией.",
		"registration_open",
		now.AddDate(0, 0, 14), now.AddDate(0, 0, 15),
		now.AddDate(0, 0, -7), now.AddDate(0, 0, 7),
	)

	upsertEvent := func(competitionID int64) int64 {
		var id int64
		scanErr := pool.QueryRow(ctx, `
			SELECT id FROM competition_event
			WHERE competition_id = $1 AND discipline_id = $2 AND age_group_id = $3
			ORDER BY id LIMIT 1
		`, competitionID, disciplineID, ageGroupID).Scan(&id)
		if scanErr == pgx.ErrNoRows {
			scanErr = pool.QueryRow(ctx, `
				INSERT INTO competition_event (competition_id, discipline_id, age_group_id)
				VALUES ($1, $2, $3) RETURNING id
			`, competitionID, disciplineID, ageGroupID).Scan(&id)
		}
		if scanErr != nil {
			log.Fatalf("upsert competition event: %v", scanErr)
		}
		return id
	}
	completedEventID := upsertEvent(completedCompetitionID)
	upcomingEventID := upsertEvent(upcomingCompetitionID)

	// Uploaded records are sufficient for API lists; physical demo files are intentionally not fabricated.
	upsertValidCertificate := func(userID int64, suffix string) {
		var fileID int64
		storageKey := "demo/anti_doping_" + suffix + ".pdf"
		if scanErr := pool.QueryRow(ctx, `
			INSERT INTO uploaded_file (
				purpose_code, storage_key, original_filename,
				content_type, size_bytes, sha256, status
			)
			VALUES ('anti_doping', $1, $2, 'application/pdf', 1024, $3, 'attached')
			ON CONFLICT (storage_key) DO UPDATE SET status = 'attached', updated_at = now()
			RETURNING id
		`, storageKey, "anti_doping_"+suffix+".pdf", fmt.Sprintf("%064x", userID)).Scan(&fileID); scanErr != nil {
			log.Fatalf("upsert anti-doping file: %v", scanErr)
		}

		completedAt := now
		validUntil := time.Date(completedAt.Year(), time.December, 31, 0, 0, 0, 0, time.UTC)
		mustExec("upsert anti-doping certificate", `
			INSERT INTO anti_doping (
				user_id, file_id, status, completed_at, valid_until,
				reviewed_by_user_id, reviewed_at, rejection_reason
			)
			VALUES ($1, $2, 'valid', $3, $4, $5, now(), NULL)
			ON CONFLICT (file_id) DO UPDATE SET
				user_id = EXCLUDED.user_id,
				status = 'valid',
				completed_at = EXCLUDED.completed_at,
				valid_until = EXCLUDED.valid_until,
				reviewed_by_user_id = EXCLUDED.reviewed_by_user_id,
				reviewed_at = now(),
				rejection_reason = NULL
		`, userID, fileID, completedAt, validUntil, adminID)
	}
	upsertValidCertificate(athleteID, "athlete_1")
	upsertValidCertificate(athleteTwoID, "athlete_2")
	upsertValidCertificate(athleteThreeID, "athlete_3")

	upsertApplication := func(eventID, athleteUserID, submittedByUserID int64, status string) int64 {
		var id int64
		if scanErr := pool.QueryRow(ctx, `
			INSERT INTO competition_application (
				event_id, athlete_user_id, submitted_by_user_id, status,
				reviewed_by_user_id, reviewed_at
			)
			VALUES (
				$1, $2, $3, $4::varchar,
				CASE WHEN $4::varchar IN ('approved', 'rejected') THEN $5::bigint END,
				CASE WHEN $4::varchar IN ('approved', 'rejected') THEN now() END
			)
			ON CONFLICT (event_id, athlete_user_id) DO UPDATE SET
				submitted_by_user_id = EXCLUDED.submitted_by_user_id,
				status = EXCLUDED.status,
				reviewed_by_user_id = EXCLUDED.reviewed_by_user_id,
				reviewed_at = EXCLUDED.reviewed_at
			RETURNING id
		`, eventID, athleteUserID, submittedByUserID, status, adminID).Scan(&id); scanErr != nil {
			log.Fatalf("upsert competition application: %v", scanErr)
		}
		return id
	}
	applicationOneID := upsertApplication(completedEventID, athleteID, coachID, "approved")
	applicationTwoID := upsertApplication(completedEventID, athleteTwoID, athleteTwoID, "approved")
	_ = upsertApplication(upcomingEventID, athleteThreeID, athleteThreeID, "pending")

	upsertCompetitionResult := func(applicationID int64, place int, score float64, awardedRankID *int64) int64 {
		var id int64
		if scanErr := pool.QueryRow(ctx, `
			INSERT INTO competition_result (
				application_id, place, score, result_status, awarded_rank_id
			)
			VALUES ($1, $2, $3, 'finished', $4)
			ON CONFLICT (application_id) DO UPDATE SET
				place = EXCLUDED.place,
				score = EXCLUDED.score,
				result_status = EXCLUDED.result_status,
				awarded_rank_id = EXCLUDED.awarded_rank_id
			RETURNING id
		`, applicationID, place, score, awardedRankID).Scan(&id); scanErr != nil {
			log.Fatalf("upsert competition result: %v", scanErr)
		}
		return id
	}
	_ = upsertCompetitionResult(applicationOneID, 1, 300, &rankID)
	resultTwoID := upsertCompetitionResult(applicationTwoID, 2, 220, nil)

	// Seed the modern result sheet used by the current result-management UI.
	var sheetID int64
	if err = pool.QueryRow(ctx, `
		INSERT INTO competition_event_result_sheet (
			event_id, mode, task_count, status, calculated_at,
			approved_by_user_id, approved_at, protocol_storage_key
		)
		VALUES ($1, 'individual', 3, 'approved', now(), $2, now(), 'demo/protocol.pdf')
		ON CONFLICT (event_id) DO UPDATE SET
			mode = 'individual', task_count = 3, status = 'approved',
			calculated_at = now(), approved_by_user_id = EXCLUDED.approved_by_user_id,
			approved_at = now(), protocol_storage_key = EXCLUDED.protocol_storage_key,
			updated_at = now(), revision = competition_event_result_sheet.revision + 1
		RETURNING id
	`, completedEventID, adminID).Scan(&sheetID); err != nil {
		log.Fatal("upsert result sheet:", err)
	}
	mustExec("reset demo result entries", `DELETE FROM competition_event_result_entry WHERE sheet_id = $1`, sheetID)

	createResultEntry := func(applicationID int64, place, solved int, totalTime, lastSolve int64, negatives int) {
		var entryID int64
		if scanErr := pool.QueryRow(ctx, `
			INSERT INTO competition_event_result_entry (
				sheet_id, result_status, solved_count, total_time_deciseconds,
				last_solve_time_deciseconds, negative_attempts, place
			)
			VALUES ($1, 'finished', $2, $3, $4, $5, $6)
			RETURNING id
		`, sheetID, solved, totalTime, lastSolve, negatives, place).Scan(&entryID); scanErr != nil {
			log.Fatalf("create result entry: %v", scanErr)
		}
		mustExec("link result entry", `
			INSERT INTO competition_event_result_entry_application (entry_id, application_id)
			VALUES ($1, $2)
		`, entryID, applicationID)

		for task := 1; task <= 3; task++ {
			status := "solved"
			var solveTime any = int64(task*18000 + place*900)
			negativeAttempts := 0
			if task > solved {
				status = "attempted"
				solveTime = nil
				negativeAttempts = 1
			}
			mustExec("create task result", `
				INSERT INTO competition_event_task_result (
					entry_id, task_number, status, solve_time_deciseconds, negative_attempts
				)
				VALUES ($1, $2, $3, $4, $5)
			`, entryID, task, status, solveTime, negativeAttempts)
		}
	}
	createResultEntry(applicationOneID, 1, 3, 111600, 54900, 1)
	createResultEntry(applicationTwoID, 2, 2, 89100, 37800, 2)
	mustExec("set event protocol", `
		UPDATE competition_event SET protocol_url = '/api/v1/files/demo/protocol.pdf'
		WHERE id = $1
	`, completedEventID)

	// Team data supports both the teams registry and team-result configuration examples.
	var teamID int64
	err = pool.QueryRow(ctx, `
		SELECT id FROM team WHERE name = 'Сборная Оренбургской области'
		ORDER BY id LIMIT 1
	`).Scan(&teamID)
	if err == pgx.ErrNoRows {
		err = pool.QueryRow(ctx, `
			INSERT INTO team (name, team_type)
			VALUES ('Сборная Оренбургской области', 'regional_squad')
			RETURNING id
		`).Scan(&teamID)
	}
	if err != nil {
		log.Fatal("upsert team:", err)
	}
	for _, userID := range []int64{athleteID, athleteTwoID} {
		mustExec("assign team member", `
			INSERT INTO team_member (team_id, user_id, joined_at)
			SELECT $1, $2, CURRENT_DATE - 365
			WHERE NOT EXISTS (
				SELECT 1 FROM team_member
				WHERE team_id = $1 AND user_id = $2 AND left_at IS NULL
			)
		`, teamID, userID)
	}
	mustExec("assign team coach", `
		INSERT INTO team_coach (team_id, coach_user_id)
		SELECT $1, $2
		WHERE NOT EXISTS (
			SELECT 1 FROM team_coach WHERE team_id = $1 AND coach_user_id = $2
		)
	`, teamID, coachID)

	for _, competitionID := range []int64{completedCompetitionID, upcomingCompetitionID} {
		mustExec("assign competition judge", `
			INSERT INTO competition_judge (competition_id, judge_user_id, position_id)
			SELECT $1, $2, $3
			WHERE NOT EXISTS (
				SELECT 1 FROM competition_judge
				WHERE competition_id = $1 AND judge_user_id = $2 AND position_id = $3
			)
		`, competitionID, judgeID, judgePositionID)
	}

	// Keep one pending rank request for the federation review queue.
	mustExec("create rank application", `
		INSERT INTO rank_application (
			athlete_user_id, rank_id, discipline_id, submitted_by_user_id,
			status, application_type, competition_id, result_id
		)
		SELECT $1, $2, $3, $4, 'pending', 'assignment', $5, $6
		WHERE NOT EXISTS (
			SELECT 1 FROM rank_application
			WHERE athlete_user_id = $1 AND rank_id = $2 AND discipline_id = $3
			  AND status = 'pending'
		)
	`, athleteTwoID, rankID, disciplineID, coachID, completedCompetitionID, resultTwoID)
	mustExec("upsert EVSK standard", `
		INSERT INTO rank_evsk_standard (
			rank_id, discipline_id, minimum_place, minimum_score, valid_from
		)
		VALUES ($1, $2, 3, 200, CURRENT_DATE - 365)
		ON CONFLICT (rank_id, discipline_id) DO UPDATE SET
			minimum_place = EXCLUDED.minimum_place,
			minimum_score = EXCLUDED.minimum_score,
			valid_from = EXCLUDED.valid_from,
			valid_to = NULL
	`, rankID, disciplineID)

	for _, deadline := range []struct {
		name         string
		deadlineType string
		dueAt        time.Time
	}{
		{"Квартальный отчёт федерации", "report", now.AddDate(0, 0, 10)},
		{"Подача документов на губернаторскую премию", "governor_award", now.AddDate(0, 1, 0)},
	} {
		mustExec("create report deadline", `
			INSERT INTO report_deadline (
				name, due_at, status, created_by_user_id, deadline_type
			)
			SELECT $1::varchar, $2::timestamptz, 'planned', $3::bigint, $4::varchar
			WHERE NOT EXISTS (
				SELECT 1 FROM report_deadline
				WHERE name = $1::varchar AND deadline_type = $4::varchar
			)
		`, deadline.name, deadline.dueAt, adminID, deadline.deadlineType)
	}

	log.Printf("Demo seed applied successfully: 6 users, 2 competitions, 3 applications, 2 results; password %q", demoPassword)
}
