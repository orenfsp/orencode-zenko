package competition

import (
	"context"
	"net/url"
	"strings"
	"unicode/utf8"

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

func (s *Service) Create(ctx context.Context, createdBy int64, req CreateRequest) (ResponseDTO, error) {
	if createdBy <= 0 {
		return ResponseDTO{}, competitionerrs.ErrForbidden
	}

	input, err := validateCreateRequest(createdBy, req)
	if err != nil {
		return ResponseDTO{}, err
	}
	if err := s.repo.ValidateReferences(ctx, CompetitionReferences{
		DisciplineID: input.DisciplineID, CompetitionTypeID: input.CompetitionTypeID,
		CompetitionLevelID: input.CompetitionLevelID, CalendarPlanID: input.CalendarPlanID,
	}); err != nil {
		return ResponseDTO{}, err
	}

	created, err := s.repo.Create(ctx, input)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(created), nil
}

func validateCreateRequest(createdBy int64, req CreateRequest) (CreateCompetition, error) {
	name := strings.TrimSpace(req.Name)
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if name == "" || utf8.RuneCountInString(name) > 250 ||
		req.DisciplineID <= 0 || req.CompetitionTypeID <= 0 || req.CompetitionLevelID <= 0 ||
		(format != "offline" && format != "online" && format != "mixed") ||
		req.StartAt.IsZero() || req.EndAt.IsZero() ||
		req.RegistrationStartsAt.IsZero() || req.RegistrationEndsAt.IsZero() ||
		!req.StartAt.Before(req.EndAt) ||
		!req.RegistrationStartsAt.Before(req.RegistrationEndsAt) ||
		req.RegistrationEndsAt.After(req.StartAt) {
		return CreateCompetition{}, competitionerrs.ErrInvalidCreate
	}
	if req.CalendarPlanID != nil && *req.CalendarPlanID <= 0 {
		return CreateCompetition{}, competitionerrs.ErrInvalidCreate
	}

	description := normalizedOptional(req.Description)
	venueAddress := normalizedOptional(req.VenueAddress)
	onlineURL := normalizedOptional(req.OnlineURL)
	calendarPlanNumber := normalizedOptional(req.CalendarPlanNumber)
	regulationURL := normalizedOptional(req.RegulationURL)

	if (format == "offline" || format == "mixed") && venueAddress == nil {
		return CreateCompetition{}, competitionerrs.ErrInvalidCreate
	}
	if (format == "online" || format == "mixed") && (onlineURL == nil || !validHTTPURL(*onlineURL)) {
		return CreateCompetition{}, competitionerrs.ErrInvalidCreate
	}
	if onlineURL != nil && !validHTTPURL(*onlineURL) {
		return CreateCompetition{}, competitionerrs.ErrInvalidCreate
	}
	if regulationURL != nil && !validHTTPURL(*regulationURL) {
		return CreateCompetition{}, competitionerrs.ErrInvalidCreate
	}
	if req.CalendarPlanID == nil && calendarPlanNumber != nil {
		return CreateCompetition{}, competitionerrs.ErrInvalidCreate
	}

	return CreateCompetition{
		Name: name, Description: description, DisciplineID: req.DisciplineID,
		CompetitionTypeID: req.CompetitionTypeID, CompetitionLevelID: req.CompetitionLevelID,
		Format: format, StartAt: req.StartAt, EndAt: req.EndAt,
		RegistrationStartsAt: req.RegistrationStartsAt, RegistrationEndsAt: req.RegistrationEndsAt,
		VenueAddress: venueAddress, OnlineURL: onlineURL, CreatedBy: createdBy,
		CalendarPlanID: req.CalendarPlanID, CalendarPlanNumber: calendarPlanNumber,
		RegulationURL: regulationURL,
	}, nil
}

func normalizedOptional(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func validHTTPURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
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
