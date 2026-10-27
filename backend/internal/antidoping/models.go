package antidoping

import "time"

const (
	StatusManualReview = "manual_review"
	StatusValid        = "valid"
	StatusExpired      = "expired"
	StatusRejected     = "rejected"
)

type Certificate struct {
	ID              int64      `json:"id"`
	UserID          int64      `json:"user_id"`
	FileID          int64      `json:"file_id"`
	Status          string     `json:"status"`
	CompletedAt     *string    `json:"completed_at,omitempty"`
	ValidUntil      *string    `json:"valid_until,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason *string    `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ReviewRequest struct {
	Status          string  `json:"status"`
	CompletedAt     *string `json:"completed_at"`
	RejectionReason *string `json:"rejection_reason"`
}

type ReviewInput struct {
	Status          string
	CompletedAt     time.Time
	RejectionReason string
}
