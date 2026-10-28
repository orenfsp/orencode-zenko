package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type dashboardService interface {
	Get(ctx context.Context, year int) (Data, error)
}

type Handler struct {
	service dashboardService
}

func NewHandler(service dashboardService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	year := time.Now().UTC().Year()
	if rawYear := r.URL.Query().Get("year"); rawYear != "" {
		parsed, err := strconv.Atoi(rawYear)
		if err != nil {
			writeError(w, http.StatusBadRequest, "year должен быть числом")
			return
		}
		year = parsed
	}
	data, err := h.service.Get(r.Context(), year)
	if err != nil {
		if errors.Is(err, ErrInvalidYear) {
			writeError(w, http.StatusBadRequest, "недопустимый год")
			return
		}
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
