package result

import (
	"encoding/json"
	"errors"
	"fmt"
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

func (h *Handler) History(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeMessage(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	items, err := h.service.repository.ListHistory(r.Context(), identity.UserID)
	if err != nil {
		handleError(w, err)
		return
	}
	if items == nil {
		items = []HistoryItem{}
	}
	writeData(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	response, err := h.service.Get(r.Context(), competitionID, eventID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, response)
}

func (h *Handler) Configure(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	var request SettingsRequest
	if err := decodeStrictJSON(r, &request); err != nil {
		writeMessage(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	response, err := h.service.Configure(r.Context(), competitionID, eventID, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, response)
}

func (h *Handler) Save(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	var request SaveResultsRequest
	if err := decodeStrictJSON(r, &request); err != nil {
		writeMessage(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	response, err := h.service.Save(r.Context(), competitionID, eventID, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, response)
}

func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	response, err := h.service.Calculate(r.Context(), competitionID, eventID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, response)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeMessage(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	response, err := h.service.Approve(r.Context(), competitionID, eventID, identity.UserID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, response)
}

func (h *Handler) DownloadProtocol(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	download, err := h.service.OpenProtocol(r.Context(), competitionID, eventID)
	if err != nil {
		handleError(w, err)
		return
	}
	defer download.Content.Close()
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%q", download.Filename))
	w.WriteHeader(http.StatusOK)
	if _, err := io.Copy(w, download.Content); err != nil && logger.Log != nil {
		logger.Log.Error("stream competition protocol", zap.Error(err))
	}
}

func resultPathIDs(r *http.Request) (int64, int64, error) {
	competitionID, err := parsePositiveID(r.PathValue("competitionID"))
	if err != nil {
		return 0, 0, err
	}
	eventID, err := parsePositiveID(r.PathValue("eventID"))
	if err != nil {
		return 0, 0, err
	}
	return competitionID, eventID, nil
}

func parsePositiveID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrInvalidRequest
	}
	return id, nil
}

func decodeStrictJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrInvalidRequest
	}
	return nil
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidRequest), errors.Is(err, ErrInvalidTeamSetup):
		writeMessage(w, http.StatusBadRequest, "некорректные данные результатов")
	case errors.Is(err, ErrEventNotFound):
		writeMessage(w, http.StatusNotFound, "мероприятие соревнования не найдено")
	case errors.Is(err, ErrNotConfigured):
		writeMessage(w, http.StatusNotFound, "таблица результатов не настроена")
	case errors.Is(err, ErrEntryNotFound):
		writeMessage(w, http.StatusNotFound, "строка результатов не найдена")
	case errors.Is(err, ErrProtocolNotFound):
		writeMessage(w, http.StatusNotFound, "протокол не найден")
	case errors.Is(err, ErrAlreadyApproved):
		writeMessage(w, http.StatusConflict, "результаты уже утверждены")
	case errors.Is(err, ErrApplicationsChanged):
		writeMessage(w, http.StatusConflict, "состав одобренных заявок изменился; повторите настройку таблицы")
	case errors.Is(err, ErrStaleResults):
		writeMessage(w, http.StatusConflict, "результаты изменились; повторите операцию")
	case errors.Is(err, ErrResetConfirmation):
		writeMessage(w, http.StatusConflict, "изменение конфигурации удалит черновик; требуется confirm_reset")
	case errors.Is(err, ErrUnsupportedDiscipline):
		writeMessage(w, http.StatusUnprocessableEntity, "автоматический расчёт для этой дисциплины пока не поддерживается")
	default:
		if logger.Log != nil {
			logger.Log.Error("competition result request failed", zap.Error(err))
		}
		writeMessage(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeMessage(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
