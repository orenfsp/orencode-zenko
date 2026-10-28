package competitionevent

import (
	"encoding/json"
	"errors"
	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"io"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListPublic(w http.ResponseWriter, r *http.Request) {
	competitionID, ok := pathID(w, r, "competitionID", "некорректный ID соревнования")
	if !ok {
		return
	}
	items, err := h.service.ListPublic(r.Context(), competitionID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func (h *Handler) GetPublic(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, ok := eventPathIDs(w, r)
	if !ok {
		return
	}
	item, err := h.service.GetPublic(r.Context(), competitionID, eventID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, item)
}

func (h *Handler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		return
	}
	competitionID, ok := pathID(w, r, "competitionID", "некорректный ID соревнования")
	if !ok {
		return
	}
	items, err := h.service.ListAdmin(r.Context(), competitionID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func (h *Handler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		return
	}
	competitionID, eventID, ok := eventPathIDs(w, r)
	if !ok {
		return
	}
	item, err := h.service.GetAdmin(r.Context(), competitionID, eventID)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, item)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		return
	}
	competitionID, ok := pathID(w, r, "competitionID", "некорректный ID соревнования")
	if !ok {
		return
	}
	var request CreateRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	created, err := h.service.Create(r.Context(), competitionID, request)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, created)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		return
	}
	competitionID, eventID, ok := eventPathIDs(w, r)
	if !ok {
		return
	}
	var request UpdateRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	updated, err := h.service.Update(r.Context(), competitionID, eventID, request)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, updated)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if !authenticated(w, r) {
		return
	}
	competitionID, eventID, ok := eventPathIDs(w, r)
	if !ok {
		return
	}
	if err := h.service.Delete(r.Context(), competitionID, eventID); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func authenticated(w http.ResponseWriter, r *http.Request) bool {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return false
	}
	return true
}

func eventPathIDs(w http.ResponseWriter, r *http.Request) (int64, int64, bool) {
	competitionID, ok := pathID(w, r, "competitionID", "некорректный ID соревнования")
	if !ok {
		return 0, 0, false
	}
	eventID, ok := pathID(w, r, "eventID", "некорректный ID зачёта")
	return competitionID, eventID, ok
}

func pathID(w http.ResponseWriter, r *http.Request, name, message string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, message)
		return 0, false
	}
	return id, true
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
	case errors.Is(err, ErrInvalidID):
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования или зачёта")
	case errors.Is(err, ErrInvalidCreate), errors.Is(err, ErrInvalidUpdate):
		writeError(w, http.StatusBadRequest, "некорректные данные зачёта")
	case errors.Is(err, ErrCompetitionMissing):
		writeError(w, http.StatusNotFound, "соревнование не найдено")
	case errors.Is(err, ErrEventMissing):
		writeError(w, http.StatusNotFound, "зачёт не найден")
	case errors.Is(err, ErrAgeGroupMissing):
		writeError(w, http.StatusBadRequest, "возрастная группа не найдена")
	case errors.Is(err, ErrNotDraft):
		httpx.WriteError(w, http.StatusConflict, "invalid_competition_status", "изменять зачёты можно только у черновика соревнования")
	case errors.Is(err, ErrHasApplications):
		writeError(w, http.StatusConflict, "зачёт нельзя изменить или удалить, пока существуют заявки")
	default:
		if logger.Log != nil {
			logger.Log.Error("competition event request failed", zap.Error(err))
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	httpx.WriteError(w, status, httpx.ErrorCode(status), message)
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(successResponse{Data: data})
}
