package rating

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	filter, err := parseFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректные параметры рейтинга")
		return
	}
	page, err := h.service.List(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректные параметры рейтинга")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": page})
}

func (h *Handler) GetSelection(w http.ResponseWriter, r *http.Request) {
	scope, err := parseSelectionScope(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректные параметры состава")
		return
	}
	selection, err := h.service.GetSelection(r.Context(), scope)
	if err != nil {
		handleSelectionError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, selection)
}

func (h *Handler) SaveSelection(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		writeError(w, http.StatusUnauthorized, "требуется авторизация")
		return
	}
	var request SaveSelectionRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	selection, err := h.service.SaveSelection(r.Context(), identity.UserID, request)
	if err != nil {
		handleSelectionError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, selection)
}

func (h *Handler) LockSelection(w http.ResponseWriter, r *http.Request) {
	selectionID, err := strconv.ParseInt(r.PathValue("selectionID"), 10, 64)
	if err != nil || selectionID <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный id состава")
		return
	}
	selection, err := h.service.LockSelection(r.Context(), selectionID)
	if err != nil {
		handleSelectionError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, selection)
}

func (h *Handler) DownloadOrder(w http.ResponseWriter, r *http.Request) {
	scope, err := parseSelectionScope(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректные параметры состава")
		return
	}
	content, err := h.service.GenerateOrder(r.Context(), scope)
	if err != nil {
		handleSelectionError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="team-order-%d.pdf"`, scope.SeasonYear))
	w.Header().Set("Content-Length", fmt.Sprint(len(content)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

func parseSelectionScope(r *http.Request) (SelectionScope, error) {
	year, err := strconv.ParseInt(r.URL.Query().Get("year"), 10, 32)
	if err != nil {
		return SelectionScope{}, ErrInvalidInput
	}
	disciplineID, err := strconv.ParseInt(r.URL.Query().Get("discipline_id"), 10, 64)
	if err != nil {
		return SelectionScope{}, ErrInvalidInput
	}
	ageGroupID, err := strconv.ParseInt(r.URL.Query().Get("age_group_id"), 10, 64)
	if err != nil {
		return SelectionScope{}, ErrInvalidInput
	}
	return SelectionScope{SeasonYear: int32(year), DisciplineID: disciplineID, AgeGroupID: ageGroupID}, nil
}

func handleSelectionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "некорректные данные состава")
	case errors.Is(err, ErrSelectionNotFound):
		writeError(w, http.StatusNotFound, "состав не найден")
	case errors.Is(err, ErrSelectionLocked):
		writeError(w, http.StatusConflict, "состав уже зафиксирован")
	default:
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func parseFilter(r *http.Request) (Filter, error) {
	year := time.Now().Year()
	if raw := r.URL.Query().Get("year"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return Filter{}, ErrInvalidInput
		}
		year = parsed
	}
	limit, offset := int64(50), int64(0)
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return Filter{}, ErrInvalidInput
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return Filter{}, ErrInvalidInput
		}
	}
	discipline, err := optionalID(r.URL.Query().Get("discipline_id"))
	if err != nil {
		return Filter{}, err
	}
	municipality, err := optionalID(r.URL.Query().Get("municipality_id"))
	if err != nil {
		return Filter{}, err
	}
	ageGroup, err := optionalID(r.URL.Query().Get("age_group_id"))
	if err != nil {
		return Filter{}, err
	}
	return Filter{Year: int32(year), DisciplineID: discipline, MunicipalityID: municipality, AgeGroupID: ageGroup, Limit: int32(limit), Offset: int32(offset)}, nil
}

func optionalID(raw string) (*int64, error) {
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return nil, ErrInvalidInput
	}
	return &value, nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
