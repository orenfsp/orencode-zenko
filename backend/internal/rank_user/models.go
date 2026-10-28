package rankuser

import (
	"errors"
	"time"
)

var (
	ErrInvalidInput = errors.New("invalid rank application input")
	ErrForbidden    = errors.New("rank application is forbidden")
	ErrEVSKNotMet   = errors.New("evsk standard is not met")
	ErrConflict     = errors.New("active rank application already exists")
	ErrNotFound     = errors.New("rank application not found")
)

type Application struct {
	ID                int64
	AthleteUserID     int64
	RankID            int64
	DisciplineID      int64
	ApplicationType   string
	CompetitionID     *int64
	ResultID          *int64
	SubmittedByUserID int64
	Status            string
	ReviewedByUserID  *int64
	ReviewedAt        *time.Time
	RejectionReason   *string
	AssignedAt        *time.Time
	ExpiresAt         *time.Time
	OrderNumber       *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type BasisValidation struct {
	ResultExists       bool
	AthleteMatches     bool
	DisciplineMatches  bool
	CompetitionMatches bool
	ResultFinished     bool
}

func (v BasisValidation) Valid() bool {
	return v.ResultExists && v.AthleteMatches && v.DisciplineMatches && v.CompetitionMatches && v.ResultFinished
}

type RankHistory struct {
	ID              int64
	UserID          int64
	RankID          int64
	RankName        string
	DisciplineID    int64
	DisciplineName  string
	ResultID        *int64
	AwardedAt       *time.Time
	ExpiresAt       *time.Time
	OrderNumber     *string
	CompetitionID   *int64
	CompetitionName *string
}
