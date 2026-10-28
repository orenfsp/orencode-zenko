package judgecategory

import (
	"errors"
	"time"
)

const (
	PermissionReview = "judge_categories.review"
	RoleJudge        = "judge"
	DefaultLimit     = int32(50)
	MaxLimit         = int32(100)
)

var (
	ErrInvalidInput    = errors.New("invalid judge category application input")
	ErrForbidden       = errors.New("judge role required")
	ErrCategoryMissing = errors.New("judge category not found")
	ErrNotFound        = errors.New("judge category application not found")
	ErrConflict        = errors.New("pending judge category application already exists")
)

type Application struct {
	ID               int64      `json:"id"`
	JudgeUserID      int64      `json:"judge_user_id"`
	JudgeName        *string    `json:"judge_name,omitempty"`
	CategoryID       int64      `json:"category_id"`
	CategoryName     *string    `json:"category_name,omitempty"`
	ApplicationType  string     `json:"application_type"`
	Status           string     `json:"status"`
	SubmittedAt      time.Time  `json:"submitted_at"`
	ReviewedByUserID *int64     `json:"reviewed_by_user_id"`
	ReviewedAt       *time.Time `json:"reviewed_at"`
	AssignedAt       *time.Time `json:"assigned_at"`
	ExpiresAt        *time.Time `json:"expires_at"`
	OrderNumber      *string    `json:"order_number"`
	RejectionReason  *string    `json:"rejection_reason"`
}

type CurrentCategory struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	CategoryID   int64      `json:"category_id"`
	CategoryName string     `json:"category_name"`
	AssignedAt   time.Time  `json:"assigned_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	OrderNumber  *string    `json:"order_number"`
}

type Assignment struct {
	ID              int64     `json:"id"`
	CompetitionID   int64     `json:"competition_id"`
	CompetitionName string    `json:"competition_name"`
	StartAt         time.Time `json:"start_at"`
	EndAt           time.Time `json:"end_at"`
	Status          string    `json:"status"`
	PositionID      int64     `json:"position_id"`
	PositionName    string    `json:"position_name"`
	AppointedAt     time.Time `json:"appointed_at"`
}

type CreateRequest struct {
	CategoryID      int64  `json:"category_id"`
	ApplicationType string `json:"application_type"`
}

type ReviewRequest struct {
	Status          string  `json:"status"`
	AssignedAt      *string `json:"assigned_at"`
	ExpiresAt       *string `json:"expires_at"`
	OrderNumber     *string `json:"order_number"`
	RejectionReason *string `json:"rejection_reason"`
}

type ReviewInput struct {
	Status          string
	AssignedAt      *time.Time
	ExpiresAt       *time.Time
	OrderNumber     *string
	RejectionReason *string
}
type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
