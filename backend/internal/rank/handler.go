package rank

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rank/errs"
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

func (h *Handler) CreateRank(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	var req CreateDTO
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	created, err := h.service.CreateRank(r.Context(), CreateDTO{
		Name: req.Name,
		TTL: req.TTL,
	})
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, created)
}

func (h *Handler) ListRank(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.GetRankList(r.Context())
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
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

func (h *Handler) GetRank(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "")
		return
	}

	item, err := h.service.GetRank(r.Context(), int32(id))
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, item)
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errs.ErrInvalidPagination):
		writeError(w, http.StatusBadRequest, "limit должен быть от 1 до 100, offset не может быть отрицательным")
	case errors.Is(err, errs.ErrInvalidCalendarPeriod):
		writeError(w, http.StatusBadRequest, "from и to должны задавать корректный период в формате RFC3339 или YYYY-MM-DD")
	case errors.Is(err, errs.ErrInvalidCalendarFilter):
		writeError(w, http.StatusBadRequest, "discipline_id и level_id должны быть положительными целыми числами")
	case errors.Is(err, errs.ErrInvalidID):
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
	case errors.Is(err, errs.ErrNotFound):
		writeError(w, http.StatusNotFound, "соревнование не найдено")
	case errors.Is(err, errs.ErrInvalidCreate), errors.Is(err, errs.ErrInvalidReference):
		writeError(w, http.StatusBadRequest, "некорректные данные соревнования")
	case errors.Is(err, errs.ErrInvalidUpdate):
		writeError(w, http.StatusBadRequest, "некорректные изменения соревнования")
	case errors.Is(err, errs.ErrNameConflict):
		writeError(w, http.StatusConflict, "соревнование с таким названием уже существует")
	case errors.Is(err, errs.ErrNotDraft):
		writeError(w, http.StatusConflict, "редактировать можно только черновик соревнования")
	case errors.Is(err, errs.ErrForbidden):
		writeError(w, http.StatusForbidden, "создание соревнования доступно только сотруднику федерации")
	default:
		if logger.Log != nil {
			logger.Log.Error("competition request failed", zap.Error(err))
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}