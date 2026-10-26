package competition

import (
	"context"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
)

const (
	DefaultLimit int32 = 50
	MaxLimit     int32 = 100
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]ResponseDTO, error) {
	if limit < 1 || limit > MaxLimit || offset < 0 {
		return nil, competitionerrs.ErrInvalidPagination
	}

	competitions, err := s.repo.ListPublished(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	result := make([]ResponseDTO, 0, len(competitions))
	for _, item := range competitions {
		result = append(result, toResponse(item))
	}
	return result, nil
}

func (s *Service) Get(ctx context.Context, id int64) (ResponseDTO, error) {
	if id <= 0 {
		return ResponseDTO{}, competitionerrs.ErrInvalidID
	}

	item, err := s.repo.GetPublished(ctx, id)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(item), nil
}

func toResponse(item Competition) ResponseDTO {
	response := ResponseDTO{
		ID: item.ID, Name: item.Name, Description: item.Description,
		Discipline:      ReferenceDTO{ID: item.DisciplineID, Name: item.DisciplineName},
		CompetitionType: ReferenceDTO{ID: item.CompetitionTypeID, Name: item.CompetitionTypeName},
		CompetitionLevel: CompetitionLevelDTO{
			ID: item.CompetitionLevelID, Code: item.CompetitionLevelCode, Name: item.CompetitionLevelName,
		},
		Format: item.Format, StartAt: item.StartAt, EndAt: item.EndAt,
		RegistrationStartsAt: item.RegistrationStartsAt, RegistrationEndsAt: item.RegistrationEndsAt,
		VenueAddress: item.VenueAddress, OnlineURL: item.OnlineURL, Status: item.Status,
		RegulationURL: item.RegulationURL,
	}
	if item.CalendarPlanID != nil {
		response.CalendarPlan = &CalendarPlanDTO{
			ID: *item.CalendarPlanID, Name: item.CalendarPlanName,
			Year: item.CalendarPlanYear, Number: item.CalendarPlanNumber,
		}
	}
	return response
}
