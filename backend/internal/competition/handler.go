package competition

import (
	"encoding/json"
	"errors"
	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"io"
	"net/http"
	"strconv"
	"time"
	_ "time/tzdata"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

var calendarLocation, calendarLocationErr = time.LoadLocation("Asia/Yekaterinburg")

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "limit и offset должны быть целыми числами")
		return
	}
	values, present := r.URL.Query()["with_meta"]
	if present {
		if len(values) != 1 || (values[0] != "true" && values[0] != "false") {
			writeError(w, http.StatusBadRequest, "with_meta должен быть true или false")
			return
		}
		if values[0] == "true" {
			page, err := h.service.ListWithMeta(r.Context(), limit, offset)
			if err != nil {
				handleServiceError(w, err)
				return
			}
			writeSuccess(w, http.StatusOK, page)
			return
		}
	}

	items, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func (h *Handler) Calendar(w http.ResponseWriter, r *http.Request) {
	filter, err := parseCalendarFilter(r)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	items, err := h.service.Calendar(r.Context(), filter)
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
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}

	var req CreateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	created, err := h.service.Create(r.Context(), identity.UserID, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, created)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
		return
	}

	var req UpdateRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	updated, err := h.service.UpdateOwned(r.Context(), identity.UserID, id, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, updated)
}

func (h *Handler) ListAdmin(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	var status *string
	if value := r.URL.Query().Get("status"); value != "" {
		status = &value
	}
	items, err := h.service.ListAdmin(r.Context(), identity.UserID, status)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
}

func (h *Handler) GetAdmin(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
		return
	}
	item, err := h.service.GetAdmin(r.Context(), identity.UserID, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, item)
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request)        { h.transition(w, r, "under_review") }
func (h *Handler) Publish(w http.ResponseWriter, r *http.Request)       { h.transition(w, r, "published") }
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request)        { h.transition(w, r, "cancelled") }
func (h *Handler) ReturnToDraft(w http.ResponseWriter, r *http.Request) { h.transition(w, r, "draft") }

func (h *Handler) transition(w http.ResponseWriter, r *http.Request, target string) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
		return
	}
	item, err := h.service.Transition(r.Context(), identity.UserID, id, target)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, item)
}

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
		return
	}
	items, err := h.service.History(r.Context(), identity.UserID, id)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, items)
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

func parseCalendarFilter(r *http.Request) (CalendarFilter, error) {
	from, err := parseCalendarBoundary(r.URL.Query().Get("from"))
	if err != nil {
		return CalendarFilter{}, competitionerrs.ErrInvalidCalendarPeriod
	}
	to, err := parseCalendarBoundary(r.URL.Query().Get("to"))
	if err != nil {
		return CalendarFilter{}, competitionerrs.ErrInvalidCalendarPeriod
	}
	disciplineID, err := parseOptionalPositiveID(r.URL.Query().Get("discipline_id"))
	if err != nil {
		return CalendarFilter{}, competitionerrs.ErrInvalidCalendarFilter
	}
	levelID, err := parseOptionalPositiveID(r.URL.Query().Get("level_id"))
	if err != nil {
		return CalendarFilter{}, competitionerrs.ErrInvalidCalendarFilter
	}
	limit, offset, err := parsePagination(r)
	if err != nil {
		return CalendarFilter{}, competitionerrs.ErrInvalidPagination
	}

	return CalendarFilter{
		From: from, To: to, DisciplineID: disciplineID, LevelID: levelID,
		Limit: limit, Offset: offset,
	}, nil
}

func parseCalendarBoundary(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, competitionerrs.ErrInvalidCalendarPeriod
	}
	if parsed, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return parsed, nil
	}
	if calendarLocationErr != nil {
		return time.Time{}, calendarLocationErr
	}
	return time.ParseInLocation(time.DateOnly, value, calendarLocation)
}

func parseOptionalPositiveID(value string) (*int64, error) {
	if value == "" {
		return nil, nil
	}
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return nil, competitionerrs.ErrInvalidCalendarFilter
	}
	return &id, nil
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
	case errors.Is(err, competitionerrs.ErrInvalidCalendarPeriod):
		writeError(w, http.StatusBadRequest, "from и to должны задавать корректный период в формате RFC3339 или YYYY-MM-DD")
	case errors.Is(err, competitionerrs.ErrInvalidCalendarFilter):
		writeError(w, http.StatusBadRequest, "discipline_id и level_id должны быть положительными целыми числами")
	case errors.Is(err, competitionerrs.ErrInvalidID):
		writeError(w, http.StatusBadRequest, "некорректный ID соревнования")
	case errors.Is(err, competitionerrs.ErrNotFound):
		writeError(w, http.StatusNotFound, "соревнование не найдено")
	case errors.Is(err, competitionerrs.ErrInvalidCreate), errors.Is(err, competitionerrs.ErrInvalidReference):
		writeError(w, http.StatusBadRequest, "некорректные данные соревнования")
	case errors.Is(err, competitionerrs.ErrInvalidUpdate):
		writeError(w, http.StatusBadRequest, "некорректные изменения соревнования")
	case errors.Is(err, competitionerrs.ErrNameConflict):
		writeError(w, http.StatusConflict, "соревнование с таким названием уже существует")
	case errors.Is(err, competitionerrs.ErrNotDraft):
		httpx.WriteError(w, http.StatusConflict, "invalid_competition_status", "редактировать можно только черновик соревнования")
	case errors.Is(err, competitionerrs.ErrInvalidTransition):
		httpx.WriteError(w, http.StatusConflict, "invalid_competition_status", "недопустимый переход статуса соревнования")
	case errors.Is(err, competitionerrs.ErrInvalidAdminFilter):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_filter", "некорректный фильтр статуса соревнования")
	case errors.Is(err, competitionerrs.ErrNotOwner):
		writeError(w, http.StatusForbidden, "соревнование принадлежит другому администратору")
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
	httpx.WriteError(w, status, httpx.ErrorCode(status), message)
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Data: data})
}
