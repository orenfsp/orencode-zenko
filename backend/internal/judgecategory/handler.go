package judgecategory

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	var request CreateRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	application, err := h.service.Create(r.Context(), userID, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusCreated, application)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	items, err := h.service.ListMine(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func (h *Handler) ListPending(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		handleError(w, err)
		return
	}
	items, err := h.service.ListPending(r.Context(), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func (h *Handler) Review(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	applicationID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || applicationID <= 0 {
		handleError(w, ErrInvalidInput)
		return
	}
	var request ReviewRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	application, err := h.service.Review(r.Context(), reviewerID, applicationID, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, application)
}

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	category, err := h.service.GetCurrent(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, category)
}

func (h *Handler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	userID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	items, err := h.service.ListAssignments(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func authenticatedUserID(r *http.Request) (int64, bool) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	return identity.UserID, ok && identity.UserID > 0
}

func parsePagination(r *http.Request) (int32, int32, error) {
	limit, offset := int64(DefaultLimit), int64(0)
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, 0, ErrInvalidInput
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, 0, ErrInvalidInput
		}
	}
	if limit < 1 || limit > int64(MaxLimit) || offset < 0 {
		return 0, 0, ErrInvalidInput
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

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "некорректные данные заявки")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "подача заявки доступна только судье")
	case errors.Is(err, ErrCategoryMissing):
		writeError(w, http.StatusNotFound, "судейская категория не найдена")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "заявка не найдена")
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "заявка уже рассматривается или рассмотрена")
	default:
		if logger.Log != nil {
			logger.Log.Error("judge category request failed", zap.Error(err))
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
