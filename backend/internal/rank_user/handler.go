package rankuser

import (
	"encoding/json"
	"errors"
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

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "некорректные данные заявки")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "недостаточно прав для подачи заявки")
	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "такая активная заявка уже существует")
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
