package main

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/orenfsp/orencode-zenko.git/config"
	postgres "github.com/orenfsp/orencode-zenko.git/pkg/db"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("config load:", err)
	}

	ctx := context.Background()
	pool := postgres.NewPostgres(ctx, cfg.Database.URL)
	defer pool.Close()

	// ========== 1. СПРАВОЧНИКИ ==========

	// Названия таблиц передаются только из фиксированного списка ниже.
	getOrCreateID := func(table, name string) int64 {
		var id int64

		query := `SELECT id FROM ` + table + ` WHERE name = $1 LIMIT 1`
		err := pool.QueryRow(ctx, query, name).Scan(&id)
		if err == nil {
			return id
		}
		if err != pgx.ErrNoRows {
			log.Fatalf("find %s: %v", table, err)
		}

		query = `INSERT INTO ` + table + ` (name) VALUES ($1) RETURNING id`
		err = pool.QueryRow(ctx, query, name).Scan(&id)
		if err != nil {
			log.Fatalf("create %s: %v", table, err)
		}
		return id
	}

	disciplineID := getOrCreateID("discipline", "Алгоритмическое программирование")
	typeID := getOrCreateID("competition_type", "Чемпионат")
	levelID := getOrCreateID("competition_level", "Региональный")
	municipalityID := getOrCreateID("municipality", "г. Оренбург")

	// ========== 2. ОРГАНИЗАЦИЯ ==========

	var orgID int64
	err = pool.QueryRow(ctx, `
		SELECT id
		FROM organization
		WHERE name = $1
		LIMIT 1
	`, "Оренбургский государственный университет").Scan(&orgID)

	if err == pgx.ErrNoRows {
		err = pool.QueryRow(ctx, `
			INSERT INTO organization (name, organization_type)
			VALUES ($1, $2)
			RETURNING id
		`, "Оренбургский государственный университет", "вуз").Scan(&orgID)
		if err != nil {
			log.Fatal("create organization:", err)
		}
	} else if err != nil {
		log.Fatal("find organization:", err)
	}

	// ========== 3. СПОРТИВНЫЙ РАЗРЯД ==========

	var rankID int64
	err = pool.QueryRow(ctx, `
		SELECT id
		FROM rank
		WHERE name = $1
		LIMIT 1
	`, "Кандидат в мастера спорта").Scan(&rankID)

	if err == pgx.ErrNoRows {
		err = pool.QueryRow(ctx, `
			INSERT INTO rank (name, ttl)
			VALUES ($1, $2)
			RETURNING id
		`, "Кандидат в мастера спорта", 3).Scan(&rankID)
		if err != nil {
			log.Fatal("create rank:", err)
		}
	} else if err != nil {
		log.Fatal("find rank:", err)
	}
	_ = rankID

	// ========== 4. ПОЛЬЗОВАТЕЛИ ==========

	hash, err := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("generate password hash:", err)
	}
	passwordHash := string(hash)

	createUser := func(first, last, birth, email string) int64 {
		var id int64
		err := pool.QueryRow(ctx, `
			INSERT INTO "user" (
				first_name,
				last_name,
				birth,
				email,
				password_hash,
				municipality_id,
				organization_id
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (email)
			DO UPDATE SET
				first_name = EXCLUDED.first_name,
				last_name = EXCLUDED.last_name
			RETURNING id
		`, first, last, birth, email, passwordHash, municipalityID, orgID).Scan(&id)
		if err != nil {
			log.Fatal("create user:", err)
		}
		return id
	}

	athleteID := createUser("Иван", "Спортсменов", "2005-05-15", "athlete@example.local")
	coachID := createUser("Петр", "Тренеров", "1980-01-01", "coach@example.local")
	_ = createUser("Сидор", "Судейкин", "1975-02-02", "judge@example.local")

	// ========== 5. СУПЕР-АДМИНИСТРАТОР ==========

	// Resolve the configured RBAC super administrator rather than assuming an email address.
	var adminID int64
	err = pool.QueryRow(ctx, `
		SELECT u.id
		FROM "user" AS u
		JOIN role_user AS ru ON ru.user_id = u.id
		JOIN "role" AS r ON r.id = ru.role_id
		WHERE r.code = 'super_admin'
		ORDER BY u.id
		LIMIT 1
	`).Scan(&adminID)
	if err != nil {
		log.Fatalf("find RBAC super administrator: %v", err)
	}

	// ========== 6. СОРЕВНОВАНИЕ ==========

	now := time.Now().UTC()
	var competitionID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO competition (
			name,
			discipline_id,
			competition_type_id,
			competition_level_id,
			format,
			start_at,
			end_at,
			registration_starts_at,
			registration_ends_at,
			created_by
		)
		VALUES (
			'Кубок Оренбургской области',
			$1, $2, $3, 'online', $4, $5, $6, $7, $8
		)
		ON CONFLICT (name)
		DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, disciplineID, typeID, levelID,
		now.Add(48*time.Hour),
		now.Add(72*time.Hour),
		now.Add(-24*time.Hour),
		now.Add(24*time.Hour),
		adminID,
	).Scan(&competitionID)
	if err != nil {
		log.Fatal("create competition:", err)
	}

	// ========== 7. СОБЫТИЕ СОРЕВНОВАНИЯ ==========

	var eventID int64
	err = pool.QueryRow(ctx, `
		SELECT id
		FROM competition_event
		WHERE competition_id = $1
		  AND discipline_id = $2
		ORDER BY id
		LIMIT 1
	`, competitionID, disciplineID).Scan(&eventID)

	if err == pgx.ErrNoRows {
		err = pool.QueryRow(ctx, `
			INSERT INTO competition_event (competition_id, discipline_id)
			VALUES ($1, $2)
			RETURNING id
		`, competitionID, disciplineID).Scan(&eventID)
		if err != nil {
			log.Fatal("create event:", err)
		}
	} else if err != nil {
		log.Fatal("find event:", err)
	}

	// ========== 8. ФАЙЛ АНТИДОПИНГОВОГО СЕРТИФИКАТА ==========

	var fileID int64
	err = pool.QueryRow(ctx, `
		SELECT id
		FROM uploaded_file
		WHERE storage_key = $1
		LIMIT 1
	`, "demo/anti_doping.pdf").Scan(&fileID)

	if err == pgx.ErrNoRows {
		err = pool.QueryRow(ctx, `
			INSERT INTO uploaded_file (
				purpose_code,
				storage_key,
				original_filename,
				content_type,
				size_bytes,
				sha256
			)
			VALUES (
				'anti_doping',
				'demo/anti_doping.pdf',
				'anti_doping.pdf',
				'application/pdf',
				1024,
				repeat('a', 64)
			)
			RETURNING id
		`).Scan(&fileID)
		if err != nil {
			log.Fatal("create uploaded_file:", err)
		}
	} else if err != nil {
		log.Fatal("find uploaded_file:", err)
	}

	// ========== 9. АНТИДОПИНГОВЫЙ СЕРТИФИКАТ ==========

	// CHECK anti_doping_review_data_check требует, чтобы valid_until
	// совпадал с 31 декабря года, указанного в completed_at.
	completedAt := time.Now().UTC()
	validUntil := time.Date(
		completedAt.Year(),
		time.December,
		31,
		0, 0, 0, 0,
		time.UTC,
	)

	// Обновляем последний сертификат спортсмена, если он уже существует.
	// rejection_reason обязательно очищается для статуса valid.
	tag, err := pool.Exec(ctx, `
		UPDATE anti_doping
		SET
			file_id = $2,
			status = 'valid',
			completed_at = $3,
			valid_until = $4,
			reviewed_by_user_id = $5,
			reviewed_at = NOW(),
			rejection_reason = NULL
		WHERE id = (
			SELECT id
			FROM anti_doping
			WHERE user_id = $1
			ORDER BY id DESC
			LIMIT 1
		)
	`, athleteID, fileID, completedAt, validUntil, adminID)
	if err != nil {
		log.Fatal("update anti-doping:", err)
	}

	if tag.RowsAffected() == 0 {
		_, err = pool.Exec(ctx, `
			INSERT INTO anti_doping (
				user_id,
				file_id,
				status,
				completed_at,
				valid_until,
				reviewed_by_user_id,
				reviewed_at,
				rejection_reason
			)
			VALUES ($1, $2, 'valid', $3, $4, $5, NOW(), NULL)
		`, athleteID, fileID, completedAt, validUntil, adminID)
		if err != nil {
			log.Fatal("create anti-doping:", err)
		}
	}

	log.Println("Demo anti-doping certificate valid until:", validUntil.Format("2006-01-02"))

	// ========== 10. ЗАЯВКА СПОРТСМЕНА ==========

	var applicationID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO competition_application (
			event_id,
			athlete_user_id,
			submitted_by_user_id,
			status
		)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (event_id, athlete_user_id)
		DO UPDATE SET status = 'pending'
		RETURNING id
	`, eventID, athleteID, coachID).Scan(&applicationID)
	if err != nil {
		log.Fatal("create application:", err)
	}

	// ========== 11. РЕЗУЛЬТАТ СОРЕВНОВАНИЯ ==========

	_, err = pool.Exec(ctx, `
		INSERT INTO competition_result (
			application_id,
			place,
			score,
			result_status
		)
		VALUES ($1, 1, 100.0, 'finished')
		ON CONFLICT (application_id) DO NOTHING
	`, applicationID)
	if err != nil {
		log.Fatal("create result:", err)
	}

	// ========== 12. ПРИВЯЗКА TELEGRAM ==========

	_, err = pool.Exec(ctx, `
		INSERT INTO telegram_links (user_id, chat_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO NOTHING
	`, athleteID, 865584637)
	if err != nil {
		log.Fatal("create Telegram link:", err)
	}

	log.Println("Demo seed applied successfully")
}
