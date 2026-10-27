package municipality

type CreateRequest struct {
	Name string `json:"name"`
}

type ResponseDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}
