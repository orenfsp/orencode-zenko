package calendarplan

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

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
	filter, err := parseListFilter(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	items, err := h.service.List(r.Context(), filter)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		handleServiceError(w, err)
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
	var request CreateRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	created, err := h.service.Create(r.Context(), request)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, created)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	var request UpdateRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	updated, err := h.service.Update(r.Context(), id, request)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, updated)
}

func (h *Handler) ListCompetitions(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	limit, offset, err := parsePagination(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	items, err := h.service.ListCompetitions(r.Context(), id, limit, offset)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func parseListFilter(r *http.Request) (ListFilter, error) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		return ListFilter{}, err
	}
	var year *int16
	if value := r.URL.Query().Get("year"); value != "" {
		parsed, err := strconv.ParseInt(value, 10, 16)
		if err != nil || parsed <= 0 {
			return ListFilter{}, ErrInvalidYear
		}
		converted := int16(parsed)
		year = &converted
	}
	return ListFilter{Year: year, Limit: limit, Offset: offset}, nil
}

func parsePagination(r *http.Request) (int32, int32, error) {
	limit, offset := int64(DefaultLimit), int64(0)
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, 0, ErrInvalidPagination
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, 0, ErrInvalidPagination
		}
	}
	if err := validatePagination(int32(limit), int32(offset)); err != nil {
		return 0, 0, err
	}
	return int32(limit), int32(offset), nil
}

func parseID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidID
	}
	return id, nil
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
	case errors.Is(err, ErrInvalidPagination):
		writeError(w, http.StatusBadRequest, "limit должен быть от 1 до 100, offset не может быть отрицательным")
	case errors.Is(err, ErrInvalidYear):
		writeError(w, http.StatusBadRequest, "year должен быть положительным числом в диапазоне smallint")
	case errors.Is(err, ErrInvalidID):
		writeError(w, http.StatusBadRequest, "некорректный ID календарного плана")
	case errors.Is(err, ErrInvalidCreate):
		writeError(w, http.StatusBadRequest, "некорректные данные календарного плана")
	case errors.Is(err, ErrInvalidUpdate):
		writeError(w, http.StatusBadRequest, "некорректные изменения календарного плана")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "календарный план не найден")
	default:
		if logger.Log != nil {
			logger.Log.Error("calendar plan request failed", zap.Error(err))
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
