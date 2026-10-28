package evsk

import (
	"encoding/json"
	"net/http"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) *Handler { return &Handler{service: s} }
func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req UpsertRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		write(w, 400, map[string]string{"message": "некорректный JSON"})
		return
	}
	item, err := h.service.Upsert(r.Context(), req)
	if err != nil {
		write(w, 400, map[string]string{"message": err.Error()})
		return
	}
	write(w, 201, map[string]any{"data": item})
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.List(r.Context())
	if err != nil {
		write(w, 500, map[string]string{"message": "внутренняя ошибка сервера"})
		return
	}
	write(w, 200, map[string]any{"data": items})
}
func write(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
