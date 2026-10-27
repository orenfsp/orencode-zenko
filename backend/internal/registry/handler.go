package registry

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	filter, err := parseUserFilter(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	page, err := h.service.ListUsers(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, page)
}

func (h *Handler) ListOrganizations(w http.ResponseWriter, r *http.Request) {
	filter, err := parseOrganizationFilter(r)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	page, err := h.service.ListOrganizations(r.Context(), filter)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeData(w, page)
}

func parseUserFilter(r *http.Request) (UserFilter, error) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		return UserFilter{}, err
	}
	filter := UserFilter{Search: queryString(r, "search"), RoleCode: queryString(r, "role"), Limit: limit, Offset: offset}
	for name, destination := range map[string]**int64{
		"discipline_id": &filter.DisciplineID, "organization_id": &filter.OrganizationID,
		"rank_id": &filter.RankID, "age_group_id": &filter.AgeGroupID, "municipality_id": &filter.MunicipalityID,
	} {
		value, err := queryID(r, name)
		if err != nil {
			return UserFilter{}, err
		}
		*destination = value
	}
	return filter, nil
}

func parseOrganizationFilter(r *http.Request) (OrganizationFilter, error) {
	limit, offset, err := parsePagination(r)
	if err != nil {
		return OrganizationFilter{}, err
	}
	municipalityID, err := queryID(r, "municipality_id")
	if err != nil {
		return OrganizationFilter{}, err
	}
	return OrganizationFilter{Search: queryString(r, "search"), MunicipalityID: municipalityID, Limit: limit, Offset: offset}, nil
}

func parsePagination(r *http.Request) (int32, int32, error) {
	limit, offset := int64(DefaultLimit), int64(0)
	var err error
	if value := r.URL.Query().Get("limit"); value != "" {
		limit, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, 0, ErrInvalidFilter
		}
	}
	if value := r.URL.Query().Get("offset"); value != "" {
		offset, err = strconv.ParseInt(value, 10, 32)
		if err != nil {
			return 0, 0, ErrInvalidFilter
		}
	}
	if !validPagination(int32(limit), int32(offset)) {
		return 0, 0, ErrInvalidFilter
	}
	return int32(limit), int32(offset), nil
}

func queryID(r *http.Request, name string) (*int64, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return nil, ErrInvalidFilter
	}
	return &parsed, nil
}

func queryString(r *http.Request, name string) *string {
	value := r.URL.Query().Get(name)
	if value == "" {
		return nil
	}
	return &value
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrInvalidFilter) {
		writeError(w, http.StatusBadRequest, "некорректные параметры реестра")
		return
	}
	if logger.Log != nil {
		logger.Log.Error("registry request failed", zap.Error(err))
	}
	writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
}

func writeData(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
