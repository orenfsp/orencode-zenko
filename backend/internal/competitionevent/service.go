package competitionevent

import "context"

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
	if request.AgeGroupID != nil && *request.AgeGroupID <= 0 {
		return ResponseDTO{}, ErrInvalidCreate
	}
	created, err := s.repo.CreateDraft(ctx, competitionID, request.AgeGroupID)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(created), nil
}

func (s *Service) Update(ctx context.Context, competitionID, eventID int64, request UpdateRequest) (ResponseDTO, error) {
	if competitionID <= 0 || eventID <= 0 {
		return ResponseDTO{}, ErrInvalidID
	}
	if !request.AgeGroupID.Present || (!request.AgeGroupID.Null && request.AgeGroupID.Value <= 0) {
		return ResponseDTO{}, ErrInvalidUpdate
	}
	var ageGroupID *int64
	if !request.AgeGroupID.Null {
		ageGroupID = &request.AgeGroupID.Value
	}
	updated, err := s.repo.UpdateDraft(ctx, competitionID, eventID, ageGroupID)
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
