package rankuser

import (
	"errors"
	"time"
)

var (
	ErrInvalidInput = errors.New("invalid rank application input")
	ErrForbidden    = errors.New("rank application is forbidden")
	ErrConflict     = errors.New("active rank application already exists")
	ErrNotFound     = errors.New("rank application not found")
)

type Application struct {
	ID                int64
	AthleteUserID     int64
	RankID            int64
	DisciplineID      int64
	SubmittedByUserID int64
	Status            string
	ReviewedByUserID  *int64
	ReviewedAt        *time.Time
	RejectionReason   *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
