package competitionapplication

import (
	"encoding/json"
	"errors"
	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	var request CreateRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	result, err := h.service.Create(r.Context(), identity.UserID, request)
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
	result, err := h.service.ListAccessible(r.Context(), identity.UserID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *Handler) GetMine(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	id, err := pathID(r)
	if err != nil {
		handleError(w, err)
		return
	}
	result, err := h.service.GetAccessible(r.Context(), identity.UserID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	id, err := pathID(r)
	if err != nil {
		handleError(w, err)
		return
	}
	result, err := h.service.Withdraw(r.Context(), identity.UserID, id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *Handler) ListForReview(w http.ResponseWriter, r *http.Request) {
	filter, err := parseListFilter(r)
	if err != nil {
		handleError(w, err)
		return
	}
	result, err := h.service.ListForReview(r.Context(), filter)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *Handler) GetForReview(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		handleError(w, err)
		return
	}
	result, err := h.service.GetForReview(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) { h.review(w, r, "approved") }
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request)  { h.review(w, r, "rejected") }

func (h *Handler) review(w http.ResponseWriter, r *http.Request, status string) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	id, err := pathID(r)
	if err != nil {
		handleError(w, err)
		return
	}
	result, err := h.service.Review(r.Context(), identity.UserID, id, status)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidInput
	}
	return id, nil
}

func parseListFilter(r *http.Request) (ListFilter, error) {
	filter := ListFilter{Limit: DefaultLimit}
	if raw := r.URL.Query().Get("event_id"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			return ListFilter{}, ErrInvalidInput
		}
		filter.EventID = &value
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("status")); raw != "" {
		filter.Status = &raw
	}
	if raw := r.URL.Query().Get("limit"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return ListFilter{}, ErrInvalidInput
		}
		filter.Limit = int32(value)
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return ListFilter{}, ErrInvalidInput
		}
		filter.Offset = int32(value)
	}
	return filter, nil
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
		writeError(w, http.StatusBadRequest, "некорректные параметры заявки")
	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "нет доступа к заявке или спортсмену")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "мероприятие или заявка не найдены")
	case errors.Is(err, ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "application_already_exists", "заявка на это мероприятие уже существует")
	case errors.Is(err, ErrRegistrationClosed):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "registration_closed", "регистрация на соревнование закрыта")
	case errors.Is(err, ErrAntiDopingRequired):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "antidoping_required", "для подачи нужен действующий антидопинговый сертификат")
	case errors.Is(err, ErrAlreadyReviewed):
		httpx.WriteError(w, http.StatusConflict, "application_already_reviewed", "заявка уже рассмотрена или недоступна для изменения")
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
