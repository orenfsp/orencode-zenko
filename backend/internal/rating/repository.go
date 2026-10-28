package rating

import "context"

type Repository interface {
	List(ctx context.Context, filter Filter) ([]Entry, error)
	DisciplineExists(ctx context.Context, disciplineID int64) (bool, error)
	AgeGroupExists(ctx context.Context, ageGroupID int64) (bool, error)
	UserHasRole(ctx context.Context, userID int64, roleCode string) (bool, error)
	GetSelection(ctx context.Context, scope SelectionScope) (Selection, error)
	SaveSelection(ctx context.Context, actorID int64, request SaveSelectionRequest) (Selection, error)
	LockSelection(ctx context.Context, selectionID int64) (Selection, error)
}
