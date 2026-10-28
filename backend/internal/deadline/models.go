package deadline

import (
	"errors"
	"time"
)

const PermissionManage = "deadlines.manage"

var (
	ErrInvalidInput = errors.New("invalid deadline input")
	ErrNotFound     = errors.New("deadline not found")
)

type Item struct {
	Type          string    `json:"type"`
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	DueAt         time.Time `json:"due_at"`
	Status        string    `json:"status"`
	UserID        *int64    `json:"user_id,omitempty"`
	UserName      *string   `json:"user_name,omitempty"`
	CompetitionID *int64    `json:"competition_id,omitempty"`
}

type Report struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	DueAt           time.Time `json:"due_at"`
	Status          string    `json:"status"`
	Type            string    `json:"type"`
	CreatedByUserID int64     `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateReportRequest struct {
	Name   string `json:"name"`
	DueAt  string `json:"due_at"`
	Status string `json:"status,omitempty"`
	Type   string `json:"type,omitempty"`
}

type UpdateReportRequest struct {
	Name   *string `json:"name"`
	DueAt  *string `json:"due_at"`
	Status *string `json:"status"`
	Type   *string `json:"type"`
}
