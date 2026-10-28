package integration

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendWebhookSignsExactBody(t *testing.T) {
	payload := []byte(`{"type":"competition.published","data":{"competition_id":7}}`)
	secret := "signing-secret-123"
	var receivedBody []byte
	var receivedSignature string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		receivedSignature = r.Header.Get("X-Zenko-Signature")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	status, err := sendWebhook(context.Background(), server.Client(), Delivery{
		ID: 1, EventType: EventCompetitionPublished, Payload: payload, URL: server.URL, SigningSecret: secret,
	})
	if err != nil || status != http.StatusNoContent {
		t.Fatalf("sendWebhook() = %d, %v", status, err)
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	wantSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if string(receivedBody) != string(payload) || receivedSignature != wantSignature {
		t.Fatalf("body/signature = %s / %s", receivedBody, receivedSignature)
	}
}
