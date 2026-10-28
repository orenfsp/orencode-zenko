package competitionevent

import (
	"context"
	"strings"
	"unicode/utf8"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListPublic(ctx context.Context, competitionID int64) ([]ResponseDTO, error) {
	if competitionID <= 0 {
		return nil, ErrInvalidID
	}
	items, err := s.repo.ListPublic(ctx, competitionID)
	if err != nil {
		return nil, err
	}
	return toResponses(items), nil
}

func (s *Service) GetPublic(ctx context.Context, competitionID, eventID int64) (ResponseDTO, error) {
	if competitionID <= 0 || eventID <= 0 {
		return ResponseDTO{}, ErrInvalidID
	}
	item, err := s.repo.GetPublic(ctx, competitionID, eventID)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(item), nil
}

func (s *Service) ListAdmin(ctx context.Context, competitionID int64) ([]ResponseDTO, error) {
	if competitionID <= 0 {
		return nil, ErrInvalidID
	}
	items, err := s.repo.ListAdmin(ctx, competitionID)
	if err != nil {
		return nil, err
	}
	return toResponses(items), nil
}

func (s *Service) GetAdmin(ctx context.Context, competitionID, eventID int64) (ResponseDTO, error) {
	if competitionID <= 0 || eventID <= 0 {
		return ResponseDTO{}, ErrInvalidID
	}
	item, err := s.repo.GetAdmin(ctx, competitionID, eventID)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(item), nil
}

func (s *Service) Create(ctx context.Context, competitionID int64, request CreateRequest) (ResponseDTO, error) {
	if competitionID <= 0 {
		return ResponseDTO{}, ErrInvalidID
	}
	name := strings.TrimSpace(request.Name)
	if name == "" || utf8.RuneCountInString(name) > 250 || request.AgeGroupID != nil && *request.AgeGroupID <= 0 {
		return ResponseDTO{}, ErrInvalidCreate
	}
	created, err := s.repo.CreateDraft(ctx, competitionID, name, request.AgeGroupID)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(created), nil
}

func (s *Service) Update(ctx context.Context, competitionID, eventID int64, request UpdateRequest) (ResponseDTO, error) {
	if competitionID <= 0 || eventID <= 0 {
		return ResponseDTO{}, ErrInvalidID
	}
	if !request.Name.Present && !request.AgeGroupID.Present {
		return ResponseDTO{}, ErrInvalidUpdate
	}
	if request.Name.Present && (request.Name.Null || strings.TrimSpace(request.Name.Value) == "" || utf8.RuneCountInString(strings.TrimSpace(request.Name.Value)) > 250) {
		return ResponseDTO{}, ErrInvalidUpdate
	}
	if request.AgeGroupID.Present && !request.AgeGroupID.Null && request.AgeGroupID.Value <= 0 {
		return ResponseDTO{}, ErrInvalidUpdate
	}
	current, err := s.repo.GetAdmin(ctx, competitionID, eventID)
	if err != nil {
		return ResponseDTO{}, err
	}
	name := current.Name
	if request.Name.Present {
		name = strings.TrimSpace(request.Name.Value)
	}
	ageGroupID := current.AgeGroupID
	if request.AgeGroupID.Present {
		if request.AgeGroupID.Null {
			ageGroupID = nil
		} else {
			value := request.AgeGroupID.Value
			ageGroupID = &value
		}
	}
	updated, err := s.repo.UpdateDraft(ctx, competitionID, eventID, name, ageGroupID)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(updated), nil
}

func (s *Service) Delete(ctx context.Context, competitionID, eventID int64) error {
	if competitionID <= 0 || eventID <= 0 {
		return ErrInvalidID
	}
	return s.repo.DeleteDraft(ctx, competitionID, eventID)
}

func toResponses(items []Event) []ResponseDTO {
	result := make([]ResponseDTO, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result
}

func toResponse(item Event) ResponseDTO {
	response := ResponseDTO{
		ID: item.ID, CompetitionID: item.CompetitionID,
		Name:        item.Name,
		Discipline:  ReferenceDTO{ID: item.DisciplineID, Name: item.DisciplineName},
		ProtocolURL: item.ProtocolURL,
	}
	if item.AgeGroupID != nil {
		response.AgeGroup = &ReferenceDTO{ID: *item.AgeGroupID}
		if item.AgeGroupName != nil {
			response.AgeGroup.Name = *item.AgeGroupName
		}
	}
	return response
}
