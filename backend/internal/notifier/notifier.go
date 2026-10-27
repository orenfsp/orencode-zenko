package notifier

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	sqlc "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

// Start запускает слушатель уведомлений.
func Start(ctx context.Context, listenDB *pgx.Conn, queries *sqlc.Queries, tgToken string) error {
	// Создаем клиент, который игнорирует проверку сертификата (для обхода VPN/прокси)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// Передаем этот клиент в Telegram API
	bot, err := tgbotapi.NewBotAPIWithClient(tgToken, tgbotapi.APIEndpoint, client)
	if err != nil {
		return fmt.Errorf("telegram bot init: %w", err)
	}

	if _, err := listenDB.Exec(ctx, "LISTEN notification_events"); err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	log.Println("telegram notifier started")

	for {
		notification, err := listenDB.WaitForNotification(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Println("notification error:", err)
			continue
		}

		if err := handleNotification(ctx, bot, queries, notification.Payload); err != nil {
			log.Println("handle error:", err)
		}
	}
}

func handleNotification(ctx context.Context, bot *tgbotapi.BotAPI, queries *sqlc.Queries, payload string) error {
	eventID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		return fmt.Errorf("bad payload %q: %w", payload, err)
	}

	// 1. Читаем событие через sqlc
	event, err := queries.GetNotificationEvent(ctx, eventID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil
		}
		return fmt.Errorf("get event: %w", err)
	}

	if event.TelegramSent {
		return nil
	}

	// 2. Ищем chat_id через sqlc
	chatID, err := queries.GetTelegramChatID(ctx, event.RecipientID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("no telegram link for user_id=%d, skip", event.RecipientID)
			return nil
		}
		return fmt.Errorf("get chat id: %w", err)
	}

	// 3. Парсим текст из JSONB
	var data struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(event.Payload, &data); err != nil {
		return fmt.Errorf("bad payload: %w", err)
	}
	if data.Text == "" {
		data.Text = "Новое уведомление"
	}

	// 4. Отправляем в Telegram
	msg := tgbotapi.NewMessage(chatID, data.Text)
	if _, err := bot.Send(msg); err != nil {
		return fmt.Errorf("telegram send: %w", err)
	}

	// 5. Отмечаем как отправленное через sqlc
	if err := queries.MarkTelegramSent(ctx, eventID); err != nil {
		return fmt.Errorf("mark sent: %w", err)
	}

	log.Printf("sent telegram to user_id=%d (event id=%d)", event.RecipientID, eventID)
	return nil
}