package rankuser

import (
	"encoding/json"
	"errors"
	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"io"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	var req CreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	result, err := h.service.Create(r.Context(), identity.UserID, req)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, result)
}

func (h *Handler) ListMine(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	result, err := h.service.ListMine(r.Context(), identity.UserID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *Handler) ListPending(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		handleError(w, err)
		return
	}
	result, err := h.service.ListPending(r.Context(), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *Handler) ListHistory(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	result, err := h.service.ListHistory(r.Context(), identity.UserID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *Handler) ListBases(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	disciplineID, err := strconv.ParseInt(r.URL.Query().Get("discipline_id"), 10, 64)
	if err != nil || disciplineID <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный discipline_id")
		return
	}
	var athleteUserID int64
	if value := r.URL.Query().Get("athlete_user_id"); value != "" {
		athleteUserID, err = strconv.ParseInt(value, 10, 64)
		if err != nil || athleteUserID <= 0 {
			writeError(w, http.StatusBadRequest, "некорректный athlete_user_id")
			return
		}
	}
	items, err := h.service.ListBases(r.Context(), identity.UserID, athleteUserID, disciplineID)
	if err != nil {
		handleError(w, err)
		return
	}
	if items == nil {
		items = []BasisResponse{}
	}
	writeSuccess(w, http.StatusOK, items)
}

func (h *Handler) Review(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	applicationID, err := strconv.ParseInt(r.PathValue("applicationID"), 10, 64)
	if err != nil || applicationID <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный ID заявки")
		return
	}
	var req ReviewRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	result, err := h.service.Review(r.Context(), identity.UserID, applicationID, req)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
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

func parsePagination(r *http.Request) (int32, int32, error) {
	limit, offset := int64(50), int64(0)
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
	if limit < 1 || limit > 100 || offset < 0 {
		return 0, 0, ErrInvalidInput
	}
	return int32(limit), int32(offset), nil
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "некорректные данные заявки")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "недостаточно прав для подачи заявки")
	case errors.Is(err, ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "application_already_exists", "такая активная заявка уже существует")
	case errors.Is(err, ErrEVSKNotMet):
		writeError(w, http.StatusUnprocessableEntity, "EVSK standard is not met")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "заявка не найдена или уже рассмотрена")
	default:
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, message string) {
	httpx.WriteError(w, status, httpx.ErrorCode(status), message)
}
