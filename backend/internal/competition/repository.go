package competition

import (
	"context"
	"time"
)

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
	CalendarPlanID       *int64
	CalendarPlanName     *string
	CalendarPlanYear     *int16
	CalendarPlanNumber   *string
	RegulationURL        *string
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

type Repository interface {
	ListPublished(ctx context.Context, limit, offset int32) ([]Competition, error)
	GetPublished(ctx context.Context, id int64) (Competition, error)
	ValidateReferences(ctx context.Context, references CompetitionReferences) error
	Create(ctx context.Context, input CreateCompetition) (Competition, error)
}
