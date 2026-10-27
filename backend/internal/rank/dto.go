package rank

import (
	"time"
)
type CreateDTO struct {
	Name string `json:"name"`
	TTL  int    `json:"ttl"`
}

type ResponseDTO struct {
	ID      int32 `json:"id"`
	Name  	string	`json:"name"`
	TTL     int		`json:"ttl"`
	CreatedAt time.Time	`json:"created_at"`
	UpdatedAt time.Time	`json:"updated_at"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}
