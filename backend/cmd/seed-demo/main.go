package main

import (
	"context"
	"log"
	"time"

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

	// Хелпер для простых справочников
	getOrCreateID := func(table, name string) int64 {
		var id int64
		err := pool.QueryRow(ctx, `SELECT id FROM `+table+` WHERE name = $1 LIMIT 1`, name).Scan(&id)
		if err == nil {
			return id
		}
		err = pool.QueryRow(ctx, `INSERT INTO `+table+` (name) VALUES ($1) RETURNING id`, name).Scan(&id)
		if err != nil {
			log.Fatalf("create %s: %v", table, err)
		}
		return id
	}

	// ========== 1. СПРАВОЧНИКИ ==========
	disciplineID := getOrCreateID("discipline", "Алгоритмическое программирование")
	typeID := getOrCreateID("competition_type", "Чемпионат")
	levelID := getOrCreateID("competition_level", "Региональный")
	municipalityID := getOrCreateID("municipality", "г. Оренбург")
	
	// Организация (требует поле organization_type)
	var orgID int64
	err = pool.QueryRow(ctx, `SELECT id FROM organization WHERE name = 'Оренбургский государственный университет' LIMIT 1`).Scan(&orgID)
	if err != nil {
		err = pool.QueryRow(ctx, `
			INSERT INTO organization (name, organization_type) 
			VALUES ('Оренбургский государственный университет', 'вуз') 
			RETURNING id`).Scan(&orgID)
		if err != nil {
			log.Fatal("create organization:", err)
		}
	}

	// Разряд (если нет, создаем)
	var rankID int64
	err = pool.QueryRow(ctx, `SELECT id FROM rank WHERE name = 'Кандидат в мастера спорта' LIMIT 1`).Scan(&rankID)
	if err != nil {
		pool.QueryRow(ctx, `INSERT INTO rank (name, ttl) VALUES ('Кандидат в мастера спорта', 3) RETURNING id`).Scan(&rankID)
	}

	// ========== 2. ПОЛЬЗОВАТЕЛИ ==========
	hash, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
	passwordHash := string(hash)

	createUser := func(first, last, birth, email string) int64 {
		var id int64
		err := pool.QueryRow(ctx, `
			INSERT INTO "user" (first_name, last_name, birth, email, password_hash, municipality_id, organization_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (email) DO UPDATE SET first_name = EXCLUDED.first_name
			RETURNING id`, first, last, birth, email, passwordHash, municipalityID, orgID).Scan(&id)
		if err != nil {
			log.Fatal("create user:", err)
		}
		return id
	}

	athleteID := createUser("Иван", "Спортсменов", "2005-05-15", "athlete@example.local")
	coachID := createUser("Петр", "Тренеров", "1980-01-01", "coach@example.local")
	_ = createUser("Сидор", "Судейкин", "1975-02-02", "judge@example.local")

	// Админ (создан rbac-seed)
	var adminID int64
	pool.QueryRow(ctx, `SELECT id FROM "user" WHERE email = 'admin@example.local'`).Scan(&adminID)

	// ========== 3. СОРЕВНОВАНИЕ И СОБЫТИЕ ==========
	now := time.Now()
	var competitionID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO competition (name, discipline_id, competition_type_id, competition_level_id, format, start_at, end_at, registration_starts_at, registration_ends_at, created_by)
		VALUES ('Кубок Оренбургской области', $1, $2, $3, 'online', $4, $5, $6, $7, $8)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`,
		disciplineID, typeID, levelID, now.Add(48*time.Hour), now.Add(72*time.Hour), now.Add(-24*time.Hour), now.Add(24*time.Hour), adminID,
	).Scan(&competitionID)
	if err != nil {
		log.Fatal("create competition:", err)
	}

	var eventID int64
	err = pool.QueryRow(ctx, `INSERT INTO competition_event (competition_id, discipline_id) VALUES ($1, $2) RETURNING id`, competitionID, disciplineID).Scan(&eventID)
	if err != nil {
		log.Fatal("create event:", err)
	}

	// ========== 4. ЗАЯВКА И РЕЗУЛЬТАТ ==========
	var applicationID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO competition_application (event_id, athlete_user_id, submitted_by_user_id, status)
		VALUES ($1, $2, $3, 'pending')
		ON CONFLICT (event_id, athlete_user_id) DO UPDATE SET status = 'pending'
		RETURNING id`, eventID, athleteID, coachID).Scan(&applicationID)
	if err != nil {
		log.Fatal("create application:", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO competition_result (application_id, place, score, result_status)
		VALUES ($1, 1, 100.0, 'finished')
		ON CONFLICT (application_id) DO NOTHING`, applicationID)
	if err != nil {
		log.Fatal("create result:", err)
	}

	// Привязываем Telegram к спортсмену (замени chat_id на свой!)
	_, _ = pool.Exec(ctx, `INSERT INTO telegram_links (user_id, chat_id) VALUES ($1, $2) ON CONFLICT (user_id) DO NOTHING`, athleteID, 865584637)

	log.Println("Demo seed applied successfully")
}