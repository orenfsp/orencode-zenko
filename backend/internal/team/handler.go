package team

import (
	"encoding/json"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"net/http"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		write(w, http.StatusBadRequest, map[string]string{"message": "некорректный JSON"})
		return
	}
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		write(w, http.StatusUnauthorized, map[string]string{"message": "требуется авторизация"})
		return
	}
	squad, err := h.service.Create(r.Context(), identity.UserID, req)
	if err != nil {
		write(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		return
	}
	write(w, http.StatusCreated, map[string]any{"data": squad})
}
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		write(w, http.StatusUnauthorized, map[string]string{"message": "требуется авторизация"})
		return
	}
	items, err := h.service.List(r.Context(), identity.UserID)
	if err != nil {
		write(w, http.StatusInternalServerError, map[string]string{"message": "внутренняя ошибка сервера"})
		return
	}
	write(w, http.StatusOK, map[string]any{"data": items})
}
func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
