package integration

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

const maxWebhookAttempts = 5

func RunWebhookWorker(ctx context.Context, repository Repository, interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Second
	}
	client := &http.Client{
		Timeout:       10 * time.Second,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error { return http.ErrUseLastResponse },
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	processWebhookBatch(ctx, repository, client)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			processWebhookBatch(ctx, repository, client)
		}
	}
}

func processWebhookBatch(ctx context.Context, repository Repository, client *http.Client) {
	deliveries, err := repository.ClaimDeliveries(ctx, 20)
	if err != nil {
		logWebhookError("claim webhook deliveries", err)
		return
	}
	for _, delivery := range deliveries {
		responseStatus, deliveryErr := sendWebhook(ctx, client, delivery)
		if deliveryErr == nil {
			if err := repository.MarkDeliverySent(ctx, delivery.ID, responseStatus); err != nil {
				logWebhookError("mark webhook sent", err)
			}
			continue
		}
		status := "pending"
		if delivery.Attempts >= maxWebhookAttempts {
			status = "failed"
		}
		nextAttempt := time.Now().UTC().Add(webhookRetryDelay(delivery.Attempts))
		var statusPointer *int32
		if responseStatus > 0 {
			statusPointer = &responseStatus
		}
		message := deliveryErr.Error()
		if len(message) > 1000 {
			message = message[:1000]
		}
		if err := repository.MarkDeliveryFailed(ctx, delivery.ID, statusPointer, message, status, nextAttempt); err != nil {
			logWebhookError("mark webhook failed", err)
		}
	}
}

func sendWebhook(ctx context.Context, client *http.Client, delivery Delivery) (int32, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, delivery.URL, bytes.NewReader(delivery.Payload))
	if err != nil {
		return 0, err
	}
	mac := hmac.New(sha256.New, []byte(delivery.SigningSecret))
	_, _ = mac.Write(delivery.Payload)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Zenko-Event-ID", strconv.FormatInt(delivery.ID, 10))
	request.Header.Set("X-Zenko-Event-Type", delivery.EventType)
	request.Header.Set("X-Zenko-Signature", "sha256="+hex.EncodeToString(mac.Sum(nil)))

	response, err := client.Do(request)
	if err != nil {
		return 0, err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return int32(response.StatusCode), fmt.Errorf("webhook returned HTTP %d", response.StatusCode)
	}
	return int32(response.StatusCode), nil
}

func webhookRetryDelay(attempt int32) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	if attempt > maxWebhookAttempts {
		attempt = maxWebhookAttempts
	}
	return time.Duration(1<<(attempt-1)) * 5 * time.Second
}

func logWebhookError(message string, err error) {
	if logger.Log != nil {
		logger.Log.Error(message, zap.Error(err))
	}
}
