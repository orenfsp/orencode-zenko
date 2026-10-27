package competition

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type JudgeHandler struct {
	service *JudgeService
}

func NewJudgeHandler(service *JudgeService) *JudgeHandler {
	return &JudgeHandler{service: service}
}

func (h *JudgeHandler) List(w http.ResponseWriter, r *http.Request) {
	competitionID, err := parseJudgeID(r.PathValue("id"))
	if err != nil {
		handleJudgeError(w, err)
		return
	}
	items, err := h.service.ListPublic(r.Context(), competitionID)
	if err != nil {
		handleJudgeError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func (h *JudgeHandler) Assign(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	competitionID, err := parseJudgeID(r.PathValue("id"))
	if err != nil {
		handleJudgeError(w, err)
		return
	}
	var request AssignJudgeRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	created, err := h.service.Assign(r.Context(), competitionID, request)
	if err != nil {
		handleJudgeError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, created)
}

func (h *JudgeHandler) Remove(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	competitionID, err := parseJudgeID(r.PathValue("id"))
	if err != nil {
		handleJudgeError(w, err)
		return
	}
	assignmentID, err := parseJudgeID(r.PathValue("assignmentId"))
	if err != nil {
		handleJudgeError(w, err)
		return
	}
	if err := h.service.Remove(r.Context(), competitionID, assignmentID); err != nil {
		handleJudgeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseJudgeID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidJudgeRequest
	}
	return id, nil
}

func handleJudgeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidJudgeRequest), errors.Is(err, ErrInvalidJudgeReference):
		writeError(w, http.StatusBadRequest, "некорректные данные назначения судьи")
	case errors.Is(err, ErrJudgeCompetitionMissing), errors.Is(err, ErrJudgeCompetitionPrivate):
		writeError(w, http.StatusNotFound, "соревнование не найдено")
	case errors.Is(err, ErrJudgeUserMissing):
		writeError(w, http.StatusNotFound, "пользователь-судья не найден")
	case errors.Is(err, ErrJudgePositionMissing):
		writeError(w, http.StatusNotFound, "судейская должность не найдена")
	case errors.Is(err, ErrJudgeAssignmentMissing):
		writeError(w, http.StatusNotFound, "назначение судьи не найдено")
	case errors.Is(err, ErrUserIsNotJudge):
		writeError(w, http.StatusBadRequest, "пользователь не имеет роли судьи")
	case errors.Is(err, ErrJudgeAssignmentExists):
		writeError(w, http.StatusConflict, "судья уже назначен на эту должность")
	case errors.Is(err, ErrJudgeCompetitionLocked):
		writeError(w, http.StatusConflict, "назначения судей можно изменять только у черновика соревнования")
	default:
		if logger.Log != nil {
			logger.Log.Error("competition judge request failed", zap.Error(err))
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}
