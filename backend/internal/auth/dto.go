package auth

type RegisterRequest struct {
	FirstName                  string `json:"first_name"`
	LastName                   string `json:"last_name"`
	MiddleName                 string `json:"middle_name,omitempty"`
	BirthDate                  string `json:"birth_date"`
	Email                      string `json:"email"`
	Password                   string `json:"password"`
	PersonalDataConsent        bool   `json:"personal_data_consent"`
	ParentalConsentDocumentURL string `json:"parental_consent_document_url,omitempty"`
	DeviceID                   string `json:"device_id"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
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
