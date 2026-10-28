package competition

import (
	"context"
	"time"
)

var publicCompetitionStatuses = []string{
	"published",
	"registration_open",
	"registration_closed",
	"ongoing",
	"results_pending",
	"completed",
}

type Competition struct {
	ID                   int64
	Name                 string
	Description          *string
	DisciplineID         int64
	DisciplineName       string
	CompetitionTypeID    int64
	CompetitionTypeName  string
	CompetitionLevelID   int64
	CompetitionLevelCode *string
	CompetitionLevelName *string
	Format               string
	StartAt              time.Time
	EndAt                time.Time
	RegistrationStartsAt time.Time
	RegistrationEndsAt   time.Time
	VenueAddress         *string
	OnlineURL            *string
	Status               string
	CreatedBy            int64
	CalendarPlanID       *int64
	CalendarPlanName     *string
	CalendarPlanYear     *int16
	CalendarPlanNumber   *string
	RegulationURL        *string
}

type StatusHistory struct {
	ID            int64     `json:"id"`
	CompetitionID int64     `json:"competition_id"`
	FromStatus    *string   `json:"from_status,omitempty"`
	ToStatus      string    `json:"to_status"`
	ChangedBy     int64     `json:"changed_by"`
	ChangedAt     time.Time `json:"changed_at"`
}

type CompetitionReferences struct {
	DisciplineID       int64
	CompetitionTypeID  int64
	CompetitionLevelID int64
	CalendarPlanID     *int64
}

type CreateCompetition struct {
	Name                 string
	Description          *string
	DisciplineID         int64
	CompetitionTypeID    int64
	CompetitionLevelID   int64
	Format               string
	StartAt              time.Time
	EndAt                time.Time
	RegistrationStartsAt time.Time
	RegistrationEndsAt   time.Time
	VenueAddress         *string
	OnlineURL            *string
	CreatedBy            int64
	CalendarPlanID       *int64
	CalendarPlanNumber   *string
	RegulationURL        *string
}

type UpdateCompetition struct {
	ID                   int64
	Name                 string
	Description          *string
	DisciplineID         int64
	CompetitionTypeID    int64
	CompetitionLevelID   int64
	Format               string
	StartAt              time.Time
	EndAt                time.Time
	RegistrationStartsAt time.Time
	RegistrationEndsAt   time.Time
	VenueAddress         *string
	OnlineURL            *string
	CalendarPlanID       *int64
	CalendarPlanNumber   *string
	RegulationURL        *string
}

type CalendarFilter struct {
	From         time.Time
	To           time.Time
	DisciplineID *int64
	LevelID      *int64
	Limit        int32
	Offset       int32
}

type Repository interface {
	ListPublic(ctx context.Context, limit, offset int32) ([]Competition, error)
	CountPublic(ctx context.Context) (int64, error)
	ListCalendar(ctx context.Context, filter CalendarFilter) ([]Competition, error)
	GetPublic(ctx context.Context, id int64) (Competition, error)
	ValidateReferences(ctx context.Context, references CompetitionReferences) error
	Create(ctx context.Context, input CreateCompetition) (Competition, error)
	GetByID(ctx context.Context, id int64) (Competition, error)
	UpdateDraft(ctx context.Context, input UpdateCompetition) (Competition, error)
}

type WorkflowRepository interface {
	Repository
	ListAdmin(ctx context.Context, createdBy int64, status *string) ([]Competition, error)
	Transition(ctx context.Context, id, actorID int64, fromStatus, toStatus string) error
	ListHistory(ctx context.Context, id int64) ([]StatusHistory, error)
}
