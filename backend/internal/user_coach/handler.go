package usercoach

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

func authenticatedUserID(r *http.Request) (int64, bool) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	return identity.UserID, ok && identity.UserID > 0
}

func (h *Handler) SearchParticipants(w http.ResponseWriter, r *http.Request) {
	coachID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	limit, offset, err := parsePagination(r)
	if err != nil {
		handleError(w, ErrInvalidPagination)
		return
	}
	var disciplineID *int64
	if raw := r.URL.Query().Get("discipline_id"); raw != "" {
		id, parseErr := strconv.ParseInt(raw, 10, 64)
		if parseErr != nil || id <= 0 {
			handleError(w, ErrInvalidInput)
			return
		}
		disciplineID = &id
	}
	result, err := h.service.SearchParticipants(r.Context(), coachID, r.URL.Query().Get("search"), disciplineID, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, result)
}

// POST /api/v1/coach/user-coaches
func (h *Handler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	coachID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	var req CreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	result, err := h.service.Create(r.Context(), coachID, req)
	if err != nil {
		handleError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, result)
}

// GET /api/v1/coach/user-coaches
func (h *Handler) ListByCoach(
	w http.ResponseWriter,
	r *http.Request,
) {
	coachID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректные limit или offset")
		return
	}

	result, err := h.service.ListByCoach(
		r.Context(),
		coachID,
		limit,
		offset,
	)
	if err != nil {
		handleError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, result)
}

// GET /api/v1/participant/coaches
func (h *Handler) ListByParticipant(
	w http.ResponseWriter,
	r *http.Request,
) {
	participantID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректные limit или offset")
		return
	}

	result, err := h.service.ListByParticipant(
		r.Context(),
		participantID,
		limit,
		offset,
	)
	if err != nil {
		handleError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, result)
}

// PATCH /api/v1/user-coaches/{id}/confirm
func (h *Handler) Confirm(
	w http.ResponseWriter,
	r *http.Request,
) {
	participantID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	relationID, err := parseRelationID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный ID связи")
		return
	}

	result, err := h.service.Confirm(
		r.Context(),
		participantID,
		relationID,
	)
	if err != nil {
		handleError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, result)
}

// PATCH /api/v1/user-coaches/{id}/end
func (h *Handler) End(
	w http.ResponseWriter,
	r *http.Request,
) {
	actorID, ok := authenticatedUserID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	relationID, err := parseRelationID(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный ID связи")
		return
	}

	result, err := h.service.End(r.Context(), actorID, relationID)
	if err != nil {
		handleError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, result)
}

func parseRelationID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidInput
	}

	return id, nil
}

func parsePagination(r *http.Request) (int32, int32, error) {
	limit := int64(DefaultLimit)
	offset := int64(0)

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

	if limit < 1 || limit > int64(MaxLimit) || offset < 0 {
		return 0, 0, ErrInvalidPagination
	}

	return int32(limit), int32(offset), nil
}

func decodeJSON(
	w http.ResponseWriter,
	r *http.Request,
	destination any,
) error {
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

type apiResponse struct {
	Data any `json:"data"`
}

type apiError struct {
	Message string `json:"message"`
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput),
		errors.Is(err, ErrInvalidPagination):
		writeError(w, http.StatusBadRequest, "некорректные параметры запроса")

	case errors.Is(err, ErrForbidden):
		writeError(w, http.StatusForbidden, "недостаточно прав для операции")

	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "связь тренера и участника не найдена или операция недоступна")

	case errors.Is(err, ErrConflict):
		writeError(w, http.StatusConflict, "активная связь с этим тренером по данной дисциплине уже существует")

	default:
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeSuccess(
	w http.ResponseWriter,
	status int,
	data any,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiResponse{Data: data})
}

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(apiError{Message: message})
}
