package integration

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/orenfsp/orencode-zenko.git/internal/httpx"
	"github.com/orenfsp/orencode-zenko.git/internal/result"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Token(w http.ResponseWriter, r *http.Request) {
	var request TokenRequest
	if err := decodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	response, err := h.service.Token(r.Context(), request)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, "invalid_client", "invalid client credentials")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	writeData(w, http.StatusOK, response)
}

func (h *Handler) CreateWebhook(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "integration authentication required")
		return
	}
	var request CreateWebhookRequest
	if err := decodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	response, err := h.service.CreateWebhook(r.Context(), identity.ClientID, request)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusCreated, response)
}

func (h *Handler) ListWebhooks(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "integration authentication required")
		return
	}
	response, err := h.service.ListWebhooks(r.Context(), identity.ClientID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeData(w, http.StatusOK, response)
}

func (h *Handler) DeleteWebhook(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "integration authentication required")
		return
	}
	id, err := strconv.ParseInt(r.PathValue("subscriptionID"), 10, 64)
	if err != nil || id <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid subscription id")
		return
	}
	if err := h.service.DeleteWebhook(r.Context(), identity.ClientID, id); err != nil {
		handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ImportResults(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "integration authentication required")
		return
	}
	competitionID, err := strconv.ParseInt(r.PathValue("competitionID"), 10, 64)
	if err != nil || competitionID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid competition id")
		return
	}
	eventID, err := strconv.ParseInt(r.PathValue("eventID"), 10, 64)
	if err != nil || eventID <= 0 {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid event id")
		return
	}
	var request ImportResultsRequest
	if err := decodeJSON(w, r, &request); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}
	response, err := h.service.ImportResults(r.Context(), identity, competitionID, eventID, request)
	if err != nil {
		handleResultError(w, err)
		return
	}
	writeData(w, http.StatusOK, response)
}

func (h *Handler) GetResults(w http.ResponseWriter, r *http.Request) {
	identity, ok := IdentityFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "integration authentication required")
		return
	}
	competitionID, eventID, err := integrationResultPathIDs(r)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid competition or event id")
		return
	}
	response, err := h.service.GetResults(r.Context(), identity, competitionID, eventID)
	if err != nil {
		handleResultError(w, err)
		return
	}
	writeData(w, http.StatusOK, response)
}

func integrationResultPathIDs(r *http.Request) (int64, int64, error) {
	competitionID, err := strconv.ParseInt(r.PathValue("competitionID"), 10, 64)
	if err != nil || competitionID <= 0 {
		return 0, 0, ErrInvalidRequest
	}
	eventID, err := strconv.ParseInt(r.PathValue("eventID"), 10, 64)
	if err != nil || eventID <= 0 {
		return 0, 0, ErrInvalidRequest
	}
	return competitionID, eventID, nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 2<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return ErrInvalidRequest
	}
	return nil
}

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid integration request")
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "resource not found")
	case errors.Is(err, ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "conflict", "resource already exists")
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func handleResultError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidRequest), errors.Is(err, result.ErrInvalidRequest), errors.Is(err, result.ErrInvalidTeamSetup):
		httpx.WriteError(w, http.StatusBadRequest, "invalid_results", "invalid result import")
	case errors.Is(err, result.ErrEventNotFound), errors.Is(err, result.ErrNotConfigured), errors.Is(err, result.ErrEntryNotFound):
		httpx.WriteError(w, http.StatusNotFound, "not_found", "event or result table not found")
	case errors.Is(err, result.ErrAlreadyApproved), errors.Is(err, result.ErrApplicationsChanged), errors.Is(err, result.ErrStaleResults):
		httpx.WriteError(w, http.StatusConflict, "result_conflict", err.Error())
	case errors.Is(err, result.ErrUnsupportedDiscipline), errors.Is(err, result.ErrTemplateInactive),
		errors.Is(err, result.ErrTemplateDiscipline), errors.Is(err, result.ErrWrongScoringMode):
		httpx.WriteError(w, http.StatusUnprocessableEntity, "unsupported_results", err.Error())
	default:
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeData(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": data})
}
