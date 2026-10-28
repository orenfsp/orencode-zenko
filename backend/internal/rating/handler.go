package rating

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
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
