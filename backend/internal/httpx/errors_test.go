package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, http.StatusForbidden, "forbidden", "нет доступа")
	if w.Code != 403 || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status/header: %d %v", w.Code, w.Header())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != "forbidden" || body["message"] != "нет доступа" {
		t.Fatalf("body: %v", body)
	}
	if _, ok := body["details"]; ok {
		t.Fatal("empty details must be omitted")
	}
}

func TestWriteErrorDetails(t *testing.T) {
	w := httptest.NewRecorder()
	WriteErrorDetails(w, 400, "validation_error", "ошибка", map[string]string{"field": "обязательное поле"})
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["details"] == nil {
		t.Fatal("details missing")
	}
}
