package deadline

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	items, err := h.service.List(r.Context(), userID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func (h *Handler) ListReports(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.ListReports(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func (h *Handler) CreateReport(w http.ResponseWriter, r *http.Request) {
	actorID, ok := userID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	var request CreateReportRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	report, err := h.service.CreateReport(r.Context(), actorID, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusCreated, report)
}

func (h *Handler) UpdateReport(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		handleError(w, ErrInvalidInput)
		return
	}
	var request UpdateReportRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	report, err := h.service.UpdateReport(r.Context(), id, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, report)
}

func userID(r *http.Request) (int64, bool) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	return identity.UserID, ok && identity.UserID > 0
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
		writeError(w, http.StatusBadRequest, "некорректные данные срока")
	case errors.Is(err, ErrNotFound):
		writeError(w, http.StatusNotFound, "срок не найден")
	default:
		if logger.Log != nil {
			logger.Log.Error("deadline request failed", zap.Error(err))
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
