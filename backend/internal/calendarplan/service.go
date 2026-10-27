package calendarplan

import (
	"context"
	"strings"
	"unicode/utf8"
)

const (
	DefaultLimit int32 = 50
	MaxLimit     int32 = 100
	MaxNameRunes       = 200
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]ResponseDTO, error) {
	if err := validateFilter(filter); err != nil {
		return nil, err
	}
	plans, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	result := make([]ResponseDTO, 0, len(plans))
	for _, plan := range plans {
		result = append(result, toResponse(plan))
	}
	return result, nil
}

func (s *Service) Get(ctx context.Context, id int64) (ResponseDTO, error) {
	if id <= 0 {
		return ResponseDTO{}, ErrInvalidID
	}
	plan, err := s.repository.Get(ctx, id)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(plan), nil
}

func (s *Service) Create(ctx context.Context, request CreateRequest) (ResponseDTO, error) {
	name := strings.TrimSpace(request.Name)
	if !validPlan(name, request.Year) {
		return ResponseDTO{}, ErrInvalidCreate
	}
	created, err := s.repository.Create(ctx, CalendarPlan{Name: name, Year: request.Year})
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(created), nil
}

func (s *Service) Update(ctx context.Context, id int64, request UpdateRequest) (ResponseDTO, error) {
	if id <= 0 {
		return ResponseDTO{}, ErrInvalidID
	}
	if request.Empty() || (request.Name.Present && request.Name.Null) || (request.Year.Present && request.Year.Null) {
		return ResponseDTO{}, ErrInvalidUpdate
	}
	current, err := s.repository.Get(ctx, id)
	if err != nil {
		return ResponseDTO{}, err
	}
	input := UpdateCalendarPlan{ID: current.ID, Name: current.Name, Year: current.Year}
	if request.Name.Present {
		input.Name = strings.TrimSpace(request.Name.Value)
	}
	if request.Year.Present {
		input.Year = request.Year.Value
	}
	if !validPlan(input.Name, input.Year) {
		return ResponseDTO{}, ErrInvalidUpdate
	}
	updated, err := s.repository.Update(ctx, input)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(updated), nil
}

func (s *Service) ListCompetitions(ctx context.Context, planID int64, limit, offset int32) ([]CompetitionDTO, error) {
	if planID <= 0 {
		return nil, ErrInvalidID
	}
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}
	if _, err := s.repository.Get(ctx, planID); err != nil {
		return nil, err
	}
	competitions, err := s.repository.ListCompetitions(ctx, planID, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]CompetitionDTO, 0, len(competitions))
	for _, competition := range competitions {
		result = append(result, CompetitionDTO{
			ID: competition.ID, Name: competition.Name,
			StartAt: competition.StartAt, EndAt: competition.EndAt, Status: competition.Status,
			Discipline:         ReferenceDTO{ID: competition.DisciplineID, Name: competition.DisciplineName},
			CalendarPlanNumber: competition.CalendarPlanNumber,
		})
	}
	return result, nil
}

func validateFilter(filter ListFilter) error {
	if err := validatePagination(filter.Limit, filter.Offset); err != nil {
		return err
	}
	if filter.Year != nil && *filter.Year <= 0 {
		return ErrInvalidYear
	}
	return nil
}

func validatePagination(limit, offset int32) error {
	if limit < 1 || limit > MaxLimit || offset < 0 {
		return ErrInvalidPagination
	}
	return nil
}

func validPlan(name string, year int16) bool {
	return name != "" && utf8.RuneCountInString(name) <= MaxNameRunes && year > 0
}

func toResponse(plan CalendarPlan) ResponseDTO {
	return ResponseDTO{ID: plan.ID, Name: plan.Name, Year: plan.Year}
}
