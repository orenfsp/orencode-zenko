package parentalconsent

import "time"

type ConsentItem struct {
	ID          int64      `json:"id"`
	UserID      int64      `json:"user_id"`
	FirstName   string     `json:"first_name"`
	LastName    string     `json:"last_name"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	SubmittedAt time.Time  `json:"submitted_at"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
	FileID      *int64     `json:"file_id,omitempty"`
}

type ListResponse struct {
	Items  []ConsentItem `json:"items"`
	Limit  int32         `json:"limit"`
	Offset int32         `json:"offset"`
}