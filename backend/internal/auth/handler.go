package auth

import (
	"encoding/json"
	"net/http"

	v "github.com/Mccalin/vera-backend/internal/validator"
	"github.com/Mccalin/vera-backend/pkg/logger"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Login godoc
// @Summary Вход в систему
// @Description Аутентификация пользователя с помощью username и password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Учётные данные"
// @Success 200 {object} LoginResponse "Успешно"
// @Failure 400 {object} ErrorResponse "Неверный JSON"
// @Failure 401 {object} ErrorResponse "Неверные учётные данные"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// Decode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Validate
	if err := v.Validator.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	
	// Login
	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		logger.Log.Error(
        "login failed",
        zap.Error(err),
    	)

		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	// Success
	writeSuccess(w, http.StatusOK, resp)
}

// Refresh godoc
// @Summary продление сессии через рефреш
// @Description рефреш токена доступа
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Учётные данные"
// @Success 200 {object} LoginResponse "Успешно"
// @Failure 400 {object} ErrorResponse "Неверный JSON"
// @Failure 401 {object} ErrorResponse "Неверные учётные данные"
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request){
	var req RefreshTokenDTO
	// Decode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Validate
	if err := v.Validator.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	resp, err := h.service.Refresh(r.Context(), RefreshTokenDTO{
		RefreshToken: req.RefreshToken,
	})
	if err != nil{
		logger.Log.Error("error", zap.Error(err))
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return 
	}
	writeSuccess(w, http.StatusOK, resp)
	
}
// Logout godoc
// @Summary удаление сессии
// @Description разлогирование путём удаления сессии из табоици сессии
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Учётные данные"
// @Success 200 {object} LoginResponse "Успешно"
// @Failure 400 {object} ErrorResponse "Неверный JSON"
// @Failure 401 {object} ErrorResponse "Неверные учётные данные"
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request){
	var req LogoutDTO
	// Decode
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	// Validate
	if err := v.Validator.Struct(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	err := h.service.Logout(r.Context(), LogoutDTO{
		RefreshToken: req.RefreshToken,
	})
	if err != nil{
		logger.Log.Error("error", zap.Error(err))
		writeError(w, http.StatusUnauthorized, "invalid or expired refresh token")
		return 
	}
	
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(
		ErrorResponse{
			Message: message,
		},
	)
}

func writeSuccess(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(
		SuccessResponse{
			Data: data,
		},
	)
}

