package dashboard

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type dashboardService interface {
	Get(ctx context.Context) (Data, error)
}

type Handler struct {
	service dashboardService
}

func NewHandler(service dashboardService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	data, err := h.service.Get(r.Context())
	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("dashboard request failed", zap.Error(err))
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
