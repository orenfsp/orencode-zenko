package notification

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct {
	repo Repository
}

func NewHandler(repo Repository) *Handler {
	return &Handler{repo: repo}
}

// ListNotifications обрабатывает GET /api/v1/notifications
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	// TODO: Когда друг сделает парсинг JWT, заменить эту строку на:
	// userID := r.Context().Value(auth.UserIDKey).(int64)
	userID := int64(1) // ВРЕМЕННАЯ ЗАГЛУШКА для демо

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit == 0 {
		limit = 20
	}

	notifications, err := h.repo.GetUserNotifications(r.Context(), userID, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// MarkAsRead обрабатывает PATCH /api/v1/notifications/{id}/read
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	// TODO: Заменить на реальный userID из токена
	userID := int64(1) // ВРЕМЕННАЯ ЗАГЛУШКА

	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid notification id", http.StatusBadRequest)
		return
	}

	if err := h.repo.MarkAsRead(r.Context(), id, userID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}