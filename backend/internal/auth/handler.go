package auth

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	autherrs "github.com/orenfsp/orencode-zenko.git/internal/auth/errs"
	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	response, err := h.service.Register(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusCreated, response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	response, err := h.service.Login(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, response)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := decodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		writeError(w, http.StatusBadRequest, "refresh_token обязателен")
		return
	}
	response, err := h.service.Refresh(r.Context(), req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, response)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := decodeJSON(w, r, &req); err != nil || strings.TrimSpace(req.RefreshToken) == "" {
		writeError(w, http.StatusBadRequest, "refresh_token обязателен")
		return
	}
	if err := h.service.Logout(r.Context(), req.RefreshToken); err != nil {
		handleServiceError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) PatchProfile(w http.ResponseWriter, r *http.Request) {
	accessToken, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok {
		writeError(w, http.StatusUnauthorized, "требуется access-токен")
		return
	}
	var req PatchProfileRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	if req.TelegramUsername == nil && req.RoleID == nil {
		writeError(w, http.StatusBadRequest, "не переданы изменения профиля")
		return
	}
	response, err := h.service.PatchProfile(r.Context(), accessToken, req)
	if err != nil {
		handleServiceError(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, response)
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

func bearerToken(header string) (string, bool) {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", false
	}
	return parts[1], true
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, autherrs.ErrConsentRequired):
		writeError(w, http.StatusBadRequest, "необходимо согласие на обработку персональных данных")
	case errors.Is(err, autherrs.ErrParentalConsentRequired):
		writeError(w, http.StatusBadRequest, "для пользователя младше 18 лет необходимо письменное разрешение родителей")
	case errors.Is(err, autherrs.ErrInvalidRegistrationData):
		writeError(w, http.StatusBadRequest, "некорректные данные")
	case errors.Is(err, autherrs.ErrTelegramUsernameInvalid):
		writeError(w, http.StatusBadRequest, "некорректный Telegram username")
	case errors.Is(err, autherrs.ErrRoleNotFound):
		writeError(w, http.StatusBadRequest, "роль не найдена")
	case errors.Is(err, autherrs.ErrEmailAlreadyExists):
		writeError(w, http.StatusConflict, "пользователь с таким email уже существует")
	case errors.Is(err, autherrs.ErrTelegramUsernameTaken):
		writeError(w, http.StatusConflict, "этот Telegram username уже используется")
	case errors.Is(err, autherrs.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "неверный email или пароль")
	case errors.Is(err, autherrs.ErrInvalidSession), errors.Is(err, autherrs.ErrAccessDenied):
		writeError(w, http.StatusUnauthorized, "недействительная или истёкшая сессия")
	default:
		logger.Log.Error("auth request failed", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(SuccessResponse{Data: data})
}
