package athletereport

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

const registryReadPermission = "registry.read"

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
	ListRolePermissions(ctx context.Context, roleID int64) ([]string, error)
}

type Handler struct {
	service     *Service
	permissions PermissionChecker
}

func NewHandler(service *Service, permissions ...PermissionChecker) *Handler {
	handler := &Handler{service: service}
	if len(permissions) > 0 {
		handler.permissions = permissions[0]
	}
	return handler
}

func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "требуется авторизация")
		return
	}
	athleteUserID := identity.UserID
	if raw := r.URL.Query().Get("athlete_user_id"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed <= 0 {
			httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "некорректный ID спортсмена")
			return
		}
		athleteUserID = parsed
	}
	if athleteUserID != identity.UserID {
		if h.permissions == nil || identity.RoleID == nil || *identity.RoleID <= 0 {
			httpx.WriteError(w, http.StatusForbidden, "forbidden", "недостаточно прав")
			return
		}
		accountAllowed, err := h.permissions.HasPermission(r.Context(), identity.UserID, registryReadPermission)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "не удалось проверить права")
			return
		}
		rolePermissions, err := h.permissions.ListRolePermissions(r.Context(), *identity.RoleID)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "не удалось проверить права")
			return
		}
		activeRoleAllowed := false
		for _, code := range rolePermissions {
			if code == registryReadPermission {
				activeRoleAllowed = true
				break
			}
		}
		if !accountAllowed || !activeRoleAllowed {
			httpx.WriteError(w, http.StatusForbidden, "forbidden", "недостаточно прав")
			return
		}
	}
	content, err := h.service.Generate(r.Context(), athleteUserID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "не удалось сформировать отчёт")
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="athlete-report-%d.pdf"`, athleteUserID))
	w.Header().Set("Content-Length", fmt.Sprint(len(content)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}
