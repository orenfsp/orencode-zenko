package reference

import "context"

type Repository interface {
	ListDisciplines(ctx context.Context) ([]NamedItem, error)
	ListCompetitionTypes(ctx context.Context) ([]NamedItem, error)
	ListCompetitionLevels(ctx context.Context) ([]CompetitionLevel, error)
	ListAgeGroups(ctx context.Context) ([]AgeGroup, error)
}
