package competition

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "limit и offset должны быть целыми числами")
		return
	}

	items, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
		return
	}

	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, item)
}

func parsePagination(r *http.Request) (int32, int32, error) {
	limit, offset := int64(DefaultLimit), int64(0)
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, 0, err
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, 0, err
		}
	}
	return int32(limit), int32(offset), nil
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, competitionerrs.ErrInvalidPagination):
		writeError(w, http.StatusBadRequest, "limit должен быть от 1 до 100, offset не может быть отрицательным")
	case errors.Is(err, competitionerrs.ErrInvalidID):
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
	case errors.Is(err, competitionerrs.ErrNotFound):
		writeError(w, http.StatusNotFound, "соревнование не найдено")
	default:
		if logger.Log != nil {
			logger.Log.Error("competition request failed", zap.Error(err))
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Data: data})
}
