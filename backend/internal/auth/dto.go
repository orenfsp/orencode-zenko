package auth

type LoginRequest struct {
	Username string `json:"user_name" validate:"required"`
	Password string `json:"password" validate:"required"`
	DeviceID string `json:"device_id" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}
type LogoutDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
