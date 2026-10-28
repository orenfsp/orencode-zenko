package competitionevent

import "context"

// Public statuses intentionally match the Competition calendar contract.
var publicCompetitionStatuses = []string{
	"published",
	"registration_open",
	"registration_closed",
	"ongoing",
	"results_pending",
	"completed",
}

type Event struct {
	ID             int64
	CompetitionID  int64
	Name           string
	DisciplineID   int64
	DisciplineName string
	AgeGroupID     *int64
	AgeGroupName   *string
	ProtocolURL    *string
}

type Repository interface {
	ListPublic(ctx context.Context, competitionID int64) ([]Event, error)
	GetPublic(ctx context.Context, competitionID, eventID int64) (Event, error)
	ListAdmin(ctx context.Context, competitionID int64) ([]Event, error)
	GetAdmin(ctx context.Context, competitionID, eventID int64) (Event, error)
	CreateDraft(ctx context.Context, competitionID int64, name string, ageGroupID *int64) (Event, error)
	UpdateDraft(ctx context.Context, competitionID, eventID int64, name string, ageGroupID *int64) (Event, error)
	DeleteDraft(ctx context.Context, competitionID, eventID int64) error
}
