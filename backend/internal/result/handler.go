package result

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"io"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

const maxProtocolBytes = 20 << 20

func (h *Handler) UploadProtocol(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxProtocolBytes+(1<<20))
	reader, err := r.MultipartReader()
	if err != nil {
		writeMessage(w, http.StatusBadRequest, "ожидается multipart/form-data с полем file")
		return
	}
	part, err := reader.NextPart()
	if err != nil || part.FormName() != "file" || part.FileName() == "" {
		writeMessage(w, http.StatusBadRequest, "необходим файл в поле file")
		return
	}
	content, err := io.ReadAll(io.LimitReader(part, maxProtocolBytes+1))
	if err != nil {
		handleError(w, ErrInvalidRequest)
		return
	}
	if len(content) > maxProtocolBytes {
		handleError(w, ErrProtocolTooLarge)
		return
	}
	if !bytes.HasPrefix(content, []byte("%PDF-")) || !bytes.Contains(content[max(0, len(content)-1024):], []byte("%%EOF")) {
		handleError(w, ErrInvalidPDF)
		return
	}
	if next, nextErr := reader.NextPart(); nextErr != io.EOF || next != nil {
		writeMessage(w, http.StatusBadRequest, "ожидается один файл")
		return
	}
	url, err := h.service.UploadProtocol(r.Context(), competitionID, eventID, content)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, map[string]string{"protocol_url": url})
}

func (h *Handler) DeleteProtocol(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	if err := h.service.DeleteProtocol(r.Context(), competitionID, eventID); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

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
	var disciplineID *int64
	if raw := r.URL.Query().Get("discipline_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			writeMessage(w, http.StatusBadRequest, "некорректный ID дисциплины")
			return
		}
		disciplineID = &id
	}
	athleteUserID := identity.UserID
	if raw := r.URL.Query().Get("athlete_user_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			writeMessage(w, http.StatusBadRequest, "некорректный ID спортсмена")
			return
		}
		athleteUserID = id
	}
	items, err := h.service.ListHistory(r.Context(), identity.UserID, athleteUserID, disciplineID)
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

func (h *Handler) SaveManual(w http.ResponseWriter, r *http.Request) {
	competitionID, eventID, err := resultPathIDs(r)
	if err != nil {
		handleError(w, err)
		return
	}
	var request SaveManualResultsRequest
	if err := decodeStrictJSON(r, &request); err != nil {
		writeMessage(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	response, err := h.service.SaveManual(r.Context(), competitionID, eventID, request)
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
	case errors.Is(err, ErrForbidden):
		writeMessage(w, http.StatusForbidden, "нет доступа к результатам спортсмена")
	case errors.Is(err, ErrEventNotFound):
		writeMessage(w, http.StatusNotFound, "мероприятие соревнования не найдено")
	case errors.Is(err, ErrNotConfigured):
		writeMessage(w, http.StatusNotFound, "таблица результатов не настроена")
	case errors.Is(err, ErrEntryNotFound):
		writeMessage(w, http.StatusNotFound, "строка результатов не найдена")
	case errors.Is(err, ErrProtocolNotFound):
		writeMessage(w, http.StatusNotFound, "протокол не найден")
	case errors.Is(err, ErrProtocolNotApproved):
		httpx.WriteError(w, http.StatusConflict, "result_not_approved", "результаты зачёта не утверждены")
	case errors.Is(err, ErrInvalidPDF):
		writeMessage(w, http.StatusBadRequest, "файл не является PDF")
	case errors.Is(err, ErrProtocolTooLarge):
		writeMessage(w, http.StatusRequestEntityTooLarge, "размер PDF превышает 20 МБ")
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
	case errors.Is(err, ErrTemplateNotFound):
		writeMessage(w, http.StatusNotFound, "шаблон результатов не найден")
	case errors.Is(err, ErrTemplateInactive), errors.Is(err, ErrTemplateDiscipline), errors.Is(err, ErrWrongScoringMode):
		writeMessage(w, http.StatusUnprocessableEntity, "шаблон или режим зачёта не подходит")
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
	httpx.WriteError(w, status, httpx.ErrorCode(status), message)
}
