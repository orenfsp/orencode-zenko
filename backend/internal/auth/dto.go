package auth

type RegisterRequest struct {
	FirstName  string
	LastName   string
	MiddleName string
	BirthDate  string
	Email      string
	Password   string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type PatchProfileRequest struct {
	TelegramUsername *string `json:"tg_username,omitempty"`
	RoleID           *int64  `json:"role_id,omitempty"`
	RefreshToken     string  `json:"refresh_token,omitempty"`
}

type PatchProfileResponse struct {
	TelegramUsername *string `json:"tg_username,omitempty"`
	RoleID           *int64  `json:"role_id,omitempty"`
	AccessToken      string  `json:"access_token,omitempty"`
	RefreshToken     string  `json:"refresh_token,omitempty"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}
