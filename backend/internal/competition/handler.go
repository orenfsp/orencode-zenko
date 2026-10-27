package competition

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

type authorizedEmployeeKey struct{}

// WithAuthorizedEmployee records the user ID only after authentication and
// the federation employee role have been verified by integration middleware.
func WithAuthorizedEmployee(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, authorizedEmployeeKey{}, userID)
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	createdBy, ok := r.Context().Value(authorizedEmployeeKey{}).(int64)
	if !ok || createdBy <= 0 {
		writeError(w, http.StatusForbidden, "создание соревнования доступно только сотруднику федерации")
		return
	}

	var req CreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	created, err := h.service.Create(r.Context(), createdBy, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, created)
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

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain one JSON object")
	}
	return nil
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, competitionerrs.ErrInvalidPagination):
		writeError(w, http.StatusBadRequest, "limit должен быть от 1 до 100, offset не может быть отрицательным")
	case errors.Is(err, competitionerrs.ErrInvalidID):
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
	case errors.Is(err, competitionerrs.ErrNotFound):
		writeError(w, http.StatusNotFound, "соревнование не найдено")
	case errors.Is(err, competitionerrs.ErrInvalidCreate), errors.Is(err, competitionerrs.ErrInvalidReference):
		writeError(w, http.StatusBadRequest, "некорректные данные соревнования")
	case errors.Is(err, competitionerrs.ErrNameConflict):
		writeError(w, http.StatusConflict, "соревнование с таким названием уже существует")
	case errors.Is(err, competitionerrs.ErrForbidden):
		writeError(w, http.StatusForbidden, "создание соревнования доступно только сотруднику федерации")
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
