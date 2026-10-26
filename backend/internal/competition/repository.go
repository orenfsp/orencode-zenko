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

type Repository interface {
	ListPublished(ctx context.Context, limit, offset int32) ([]Competition, error)
	GetPublished(ctx context.Context, id int64) (Competition, error)
}
