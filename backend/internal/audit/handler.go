package audit

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	auditerrs "github.com/orenfsp/orencode-zenko.git/internal/audit/errs"
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
	limit, offset, err := parsePage(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "limit and offset must be integers")
		return
	}

	items, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		if errors.Is(err, auditerrs.ErrInvalidRequest) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		logger.Log.Error("list audit events failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func parsePage(r *http.Request) (int32, int32, error) {
	limit, offset := int64(50), int64(0)
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
