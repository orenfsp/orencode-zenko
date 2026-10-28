package notification

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// ListNotifications обрабатывает GET /api/v1/notifications
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	userID := identity.UserID

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit == 0 {
		limit = 20
	}

	notifications, err := h.repo.GetUserNotifications(r.Context(), identity.UserID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	// Маппим в DTO
	resp := make([]NotificationResponse, len(notifications))
	for i, n := range notifications {
		resp[i] = NotificationResponse{
			ID:        n.ID,
			Title:     n.Title,
			Message:   n.Message,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt.Time,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"data": resp})
}

// MarkAsRead обрабатывает PATCH /api/v1/notifications/{id}/read
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	identity, ok := rbac.IdentityFromContext(r.Context())
	if !ok || identity.UserID <= 0 {
<<<<<<< HEAD
		writeError(w, http.StatusUnauthorized, "требуется access-токен")
		return
	}
=======
		http.Error(w, "authentication required", http.StatusUnauthorized)
		return
	}
	userID := identity.UserID
>>>>>>> 67dca8e (fix backside menu and profile)

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid notification id", http.StatusBadRequest)
		return
	}

	if err := h.repo.MarkAsRead(r.Context(), id, identity.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
<<<<<<< HEAD

func parsePagination(r *http.Request) (int32, int32, error) {
	limit, offset := int64(20), int64(0)
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return 0, 0, err
		}
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return 0, 0, err
		}
	}
	if limit < 1 || limit > 100 || offset < 0 {
		return 0, 0, strconv.ErrRange
	}
	return int32(limit), int32(offset), nil
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": message})
}
=======
>>>>>>> 67dca8e (fix backside menu and profile)
