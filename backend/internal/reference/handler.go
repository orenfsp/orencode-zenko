package reference

import (
	"encoding/json"
	"net/http"

	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListDisciplines(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListDisciplines(r.Context())
	writeListResponse(w, items, err)
}

func (h *Handler) ListCompetitionTypes(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListCompetitionTypes(r.Context())
	writeListResponse(w, items, err)
}

func (h *Handler) ListCompetitionLevels(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListCompetitionLevels(r.Context())
	writeListResponse(w, items, err)
}

func (h *Handler) ListAgeGroups(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListAgeGroups(r.Context())
	writeListResponse(w, items, err)
}

func writeListResponse(w http.ResponseWriter, data any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		if logger.Log != nil {
			logger.Log.Error("reference data request failed", zap.Error(err))
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Message: "внутренняя ошибка сервера"})
		return
	}
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Data: data})
}
