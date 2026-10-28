package competitionapplication

import (
	"errors"
	"time"
)

const (
	PermissionReview = "competition_applications.review"
	DefaultLimit     = int32(50)
	MaxLimit         = int32(100)
)

var (
	ErrInvalidInput       = errors.New("invalid competition application input")
	ErrNotFound           = errors.New("competition application not found")
	ErrForbidden          = errors.New("competition application access forbidden")
	ErrConflict           = errors.New("competition application already exists")
	ErrRegistrationClosed = errors.New("competition registration is closed")
	ErrAntiDopingRequired = errors.New("valid anti-doping certificate is required")
	ErrAlreadyReviewed    = errors.New("competition application already reviewed")
)

type EventContext struct {
	EventID              int64
	DisciplineID         int64
	CompetitionID        int64
	CompetitionStatus    string
	RegistrationStartsAt time.Time
	RegistrationEndsAt   time.Time
}

type Application struct {
	ID                   int64      `json:"id"`
	EventID              int64      `json:"event_id"`
	AthleteUserID        int64      `json:"athlete_user_id"`
	SubmittedByUserID    int64      `json:"submitted_by_user_id"`
	Status               string     `json:"status"`
	SubmittedAt          time.Time  `json:"submitted_at"`
	ReviewedByUserID     *int64     `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt           *time.Time `json:"reviewed_at,omitempty"`
	CompetitionID        int64      `json:"competition_id"`
	CompetitionName      string     `json:"competition_name"`
	CompetitionStartAt   time.Time  `json:"competition_start_at"`
	CompetitionEndAt     time.Time  `json:"competition_end_at"`
	RegistrationStartsAt time.Time  `json:"registration_starts_at"`
	RegistrationEndsAt   time.Time  `json:"registration_ends_at"`
	DisciplineID         int64      `json:"discipline_id"`
	DisciplineName       string     `json:"discipline_name"`
	AgeGroupID           *int64     `json:"age_group_id,omitempty"`
	AgeGroupName         *string    `json:"age_group_name,omitempty"`
	AthleteFirstName     string     `json:"athlete_first_name"`
	AthleteLastName      string     `json:"athlete_last_name"`
	AthleteMiddleName    *string    `json:"athlete_middle_name,omitempty"`
	AthleteEmail         string     `json:"athlete_email"`
	SubmitterFirstName   string     `json:"submitter_first_name"`
	SubmitterLastName    string     `json:"submitter_last_name"`
	SubmitterMiddleName  *string    `json:"submitter_middle_name,omitempty"`
	CanWithdraw          bool       `json:"can_withdraw"`
}

type CreateRequest struct {
	EventID       int64 `json:"event_id"`
	AthleteUserID int64 `json:"athlete_user_id,omitempty"`
}

type ListFilter struct {
	EventID *int64
	Status  *string
	Limit   int32
	Offset  int32
}
