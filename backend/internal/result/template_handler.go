package result

import (
	"net/http"
	"strconv"
)

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var request CreateTemplateRequest
	if decodeStrictJSON(r, &request) != nil {
		writeMessage(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	item, err := h.service.CreateTemplate(r.Context(), request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusCreated, item)
}

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.URL.Query().Get("discipline_id"), 10, 64)
	if err != nil || id <= 0 {
		handleError(w, ErrInvalidRequest)
		return
	}
	items, err := h.service.ListTemplates(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, items)
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		handleError(w, err)
		return
	}
	item, err := h.service.GetTemplate(r.Context(), id)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, item)
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	id, err := parsePositiveID(r.PathValue("id"))
	if err != nil {
		handleError(w, err)
		return
	}
	var request UpdateTemplateRequest
	if decodeStrictJSON(r, &request) != nil {
		writeMessage(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	item, err := h.service.UpdateTemplate(r.Context(), id, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, item)
}
