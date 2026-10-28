package rating

import (
	"context"
	"time"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, filter Filter) (Page, error) {
	if filter.Year < 2000 || filter.Year > int32(time.Now().Year()+1) || filter.Limit < 1 || filter.Limit > 100 || filter.Offset < 0 {
		return Page{}, ErrInvalidInput
	}
	items, err := s.repository.List(ctx, filter)
	if err != nil {
		return Page{}, err
	}
	if items == nil {
		items = make([]Entry, 0)
	}
	return Page{Items: items, Limit: filter.Limit, Offset: filter.Offset}, nil
}

func (s *Service) GetSelection(ctx context.Context, scope SelectionScope) (Selection, error) {
	if !validScope(scope) {
		return Selection{}, ErrInvalidInput
	}
	return s.repository.GetSelection(ctx, scope)
}

func (s *Service) SaveSelection(ctx context.Context, actorID int64, request SaveSelectionRequest) (Selection, error) {
	if actorID <= 0 || !validScope(SelectionScope{
		SeasonYear: request.SeasonYear, DisciplineID: request.DisciplineID, AgeGroupID: request.AgeGroupID,
	}) || len(request.MemberIDs) == 0 || len(request.MemberIDs) > 100 {
		return Selection{}, ErrInvalidInput
	}

	members := make(map[int64]struct{}, len(request.MemberIDs))
	for _, memberID := range request.MemberIDs {
		if memberID <= 0 {
			return Selection{}, ErrInvalidInput
		}
		if _, duplicate := members[memberID]; duplicate {
			return Selection{}, ErrInvalidInput
		}
		members[memberID] = struct{}{}
	}
	if request.CaptainUserID != nil {
		if _, exists := members[*request.CaptainUserID]; !exists {
			return Selection{}, ErrInvalidInput
		}
	}

	disciplineExists, err := s.repository.DisciplineExists(ctx, request.DisciplineID)
	if err != nil {
		return Selection{}, err
	}
	ageGroupExists, err := s.repository.AgeGroupExists(ctx, request.AgeGroupID)
	if err != nil {
		return Selection{}, err
	}
	if !disciplineExists || !ageGroupExists {
		return Selection{}, ErrInvalidInput
	}
	for _, memberID := range request.MemberIDs {
		participant, err := s.repository.UserHasRole(ctx, memberID, "participant")
		if err != nil {
			return Selection{}, err
		}
		if !participant {
			return Selection{}, ErrInvalidInput
		}
	}
	if request.CoachUserID != nil {
		coach, err := s.repository.UserHasRole(ctx, *request.CoachUserID, "coach")
		if err != nil {
			return Selection{}, err
		}
		if !coach {
			return Selection{}, ErrInvalidInput
		}
	}
	return s.repository.SaveSelection(ctx, actorID, request)
}

func (s *Service) LockSelection(ctx context.Context, selectionID int64) (Selection, error) {
	if selectionID <= 0 {
		return Selection{}, ErrInvalidInput
	}
	return s.repository.LockSelection(ctx, selectionID)
}

func validScope(scope SelectionScope) bool {
	return scope.SeasonYear >= 2000 && scope.SeasonYear <= int32(time.Now().Year()+1) &&
		scope.DisciplineID > 0 && scope.AgeGroupID > 0
}
