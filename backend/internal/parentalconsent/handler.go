package parentalconsent

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Handler struct{ repo *Repository }

func NewHandler(repo *Repository) *Handler { return &Handler{repo: repo} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit == 0 { limit = 50 }

	items, err := h.repo.List(r.Context(), status, int32(limit), int32(offset))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ListResponse{Items: items, Limit: int32(limit), Offset: int32(offset)})
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	var body struct{ Status string `json:"status"` }
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.Status != "approved" && body.Status != "rejected") {
		http.Error(w, "status must be approved or rejected", http.StatusBadRequest)
		return
	}
	if err := h.repo.UpdateStatus(r.Context(), id, body.Status); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}