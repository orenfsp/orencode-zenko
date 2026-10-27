package rbac

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var request CreateRoleRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	role, err := h.service.CreateRole(r.Context(), request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusCreated, role)
}

func (h *Handler) AssignRole(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(r.PathValue("userID"), 10, 64)
	if err != nil || userID <= 0 {
		writeError(w, http.StatusBadRequest, "некорректный userID")
		return
	}
	var request AssignRoleRequest
	if err := decodeJSON(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	if err := h.service.AssignRole(r.Context(), userID, request.RoleID); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListRoles(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, roles)
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	permissions, err := h.service.ListPermissions(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, permissions)
}

func (h *Handler) ListSelfAssignableRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.service.ListSelfAssignableRoles(r.Context())
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, roles)
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
	case errors.Is(err, ErrInvalidRole), errors.Is(err, ErrPermissionNotFound):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrRoleAlreadyExists):
		writeError(w, http.StatusConflict, "роль с таким кодом или названием уже существует")
	case errors.Is(err, ErrRoleNotFound), errors.Is(err, ErrUserNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}
