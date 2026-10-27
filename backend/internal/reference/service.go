package reference

import "context"

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListDisciplines(ctx context.Context) ([]NamedItemDTO, error) {
	items, err := s.repository.ListDisciplines(ctx)
	if err != nil {
		return nil, err
	}
	return namedItemsToDTO(items), nil
}

func (s *Service) ListCompetitionTypes(ctx context.Context) ([]NamedItemDTO, error) {
	items, err := s.repository.ListCompetitionTypes(ctx)
	if err != nil {
		return nil, err
	}
	return namedItemsToDTO(items), nil
}

func (s *Service) ListCompetitionLevels(ctx context.Context) ([]CompetitionLevelDTO, error) {
	items, err := s.repository.ListCompetitionLevels(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]CompetitionLevelDTO, 0, len(items))
	for _, item := range items {
		result = append(result, CompetitionLevelDTO{ID: item.ID, Code: item.Code, Name: item.Name})
	}
	return result, nil
}

func (s *Service) ListAgeGroups(ctx context.Context) ([]AgeGroupDTO, error) {
	items, err := s.repository.ListAgeGroups(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]AgeGroupDTO, 0, len(items))
	for _, item := range items {
		result = append(result, AgeGroupDTO{
			ID: item.ID, Name: item.Name, MinAge: item.MinAge, MaxAge: item.MaxAge,
		})
	}
	return result, nil
}

func namedItemsToDTO(items []NamedItem) []NamedItemDTO {
	result := make([]NamedItemDTO, 0, len(items))
	for _, item := range items {
		result = append(result, NamedItemDTO{ID: item.ID, Name: item.Name})
	}
	return result
}
