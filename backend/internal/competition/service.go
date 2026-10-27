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

func (s *Service) Update(ctx context.Context, id int64, req UpdateRequest) (ResponseDTO, error) {
	if id <= 0 {
		return ResponseDTO{}, competitionerrs.ErrInvalidID
	}
	if req.Empty() {
		return ResponseDTO{}, competitionerrs.ErrInvalidUpdate
	}

	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ResponseDTO{}, err
	}
	if current.Status != "draft" {
		return ResponseDTO{}, competitionerrs.ErrNotDraft
	}

	input, err := mergeUpdate(current, req)
	if err != nil {
		return ResponseDTO{}, err
	}
	if err := s.repo.ValidateReferences(ctx, CompetitionReferences{
		DisciplineID: input.DisciplineID, CompetitionTypeID: input.CompetitionTypeID,
		CompetitionLevelID: input.CompetitionLevelID, CalendarPlanID: input.CalendarPlanID,
	}); err != nil {
		return ResponseDTO{}, err
	}

	updated, err := s.repo.UpdateDraft(ctx, input)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(updated), nil
}

func mergeUpdate(current Competition, req UpdateRequest) (UpdateCompetition, error) {
	input := UpdateCompetition{
		ID: current.ID, Name: current.Name, Description: current.Description,
		DisciplineID: current.DisciplineID, CompetitionTypeID: current.CompetitionTypeID,
		CompetitionLevelID: current.CompetitionLevelID, Format: current.Format,
		StartAt: current.StartAt, EndAt: current.EndAt,
		RegistrationStartsAt: current.RegistrationStartsAt, RegistrationEndsAt: current.RegistrationEndsAt,
		VenueAddress: current.VenueAddress, OnlineURL: current.OnlineURL,
		CalendarPlanID: current.CalendarPlanID, CalendarPlanNumber: current.CalendarPlanNumber,
		RegulationURL: current.RegulationURL,
	}

	if requiredPatchIsNull(req) {
		return UpdateCompetition{}, competitionerrs.ErrInvalidUpdate
	}
	if req.Name.Present {
		input.Name = strings.TrimSpace(req.Name.Value)
	}
	if req.Description.Present {
		input.Description = patchOptionalString(req.Description)
	}
	if req.DisciplineID.Present {
		input.DisciplineID = req.DisciplineID.Value
	}
	if req.CompetitionTypeID.Present {
		input.CompetitionTypeID = req.CompetitionTypeID.Value
	}
	if req.CompetitionLevelID.Present {
		input.CompetitionLevelID = req.CompetitionLevelID.Value
	}
	if req.Format.Present {
		input.Format = strings.ToLower(strings.TrimSpace(req.Format.Value))
	}
	if req.StartAt.Present {
		input.StartAt = req.StartAt.Value
	}
	if req.EndAt.Present {
		input.EndAt = req.EndAt.Value
	}
	if req.RegistrationStartsAt.Present {
		input.RegistrationStartsAt = req.RegistrationStartsAt.Value
	}
	if req.RegistrationEndsAt.Present {
		input.RegistrationEndsAt = req.RegistrationEndsAt.Value
	}
	if req.VenueAddress.Present {
		input.VenueAddress = patchOptionalString(req.VenueAddress)
	}
	if req.OnlineURL.Present {
		input.OnlineURL = patchOptionalString(req.OnlineURL)
	}
	if req.CalendarPlanID.Present {
		if req.CalendarPlanID.Null {
			input.CalendarPlanID = nil
		} else {
			input.CalendarPlanID = &req.CalendarPlanID.Value
		}
	}
	if req.CalendarPlanNumber.Present {
		input.CalendarPlanNumber = patchOptionalString(req.CalendarPlanNumber)
	}
	if req.RegulationURL.Present {
		input.RegulationURL = patchOptionalString(req.RegulationURL)
	}

	if !validUpdate(input) {
		return UpdateCompetition{}, competitionerrs.ErrInvalidUpdate
	}
	return input, nil
}

func requiredPatchIsNull(req UpdateRequest) bool {
	return (req.Name.Present && req.Name.Null) ||
		(req.DisciplineID.Present && req.DisciplineID.Null) ||
		(req.CompetitionTypeID.Present && req.CompetitionTypeID.Null) ||
		(req.CompetitionLevelID.Present && req.CompetitionLevelID.Null) ||
		(req.Format.Present && req.Format.Null) ||
		(req.StartAt.Present && req.StartAt.Null) ||
		(req.EndAt.Present && req.EndAt.Null) ||
		(req.RegistrationStartsAt.Present && req.RegistrationStartsAt.Null) ||
		(req.RegistrationEndsAt.Present && req.RegistrationEndsAt.Null)
}

func patchOptionalString(field PatchField[string]) *string {
	if field.Null {
		return nil
	}
	return normalizedOptional(&field.Value)
}

func validUpdate(input UpdateCompetition) bool {
	if input.Name == "" || utf8.RuneCountInString(input.Name) > 250 ||
		input.DisciplineID <= 0 || input.CompetitionTypeID <= 0 || input.CompetitionLevelID <= 0 ||
		(input.Format != "offline" && input.Format != "online" && input.Format != "mixed") ||
		input.StartAt.IsZero() || input.EndAt.IsZero() ||
		input.RegistrationStartsAt.IsZero() || input.RegistrationEndsAt.IsZero() ||
		!input.StartAt.Before(input.EndAt) ||
		!input.RegistrationStartsAt.Before(input.RegistrationEndsAt) ||
		input.RegistrationEndsAt.After(input.StartAt) ||
		(input.CalendarPlanID != nil && *input.CalendarPlanID <= 0) {
		return false
	}
	if (input.Format == "offline" || input.Format == "mixed") && input.VenueAddress == nil {
		return false
	}
	if (input.Format == "online" || input.Format == "mixed") && input.OnlineURL == nil {
		return false
	}
	if input.OnlineURL != nil && !validHTTPURL(*input.OnlineURL) {
		return false
	}
	if input.RegulationURL != nil && !validHTTPURL(*input.RegulationURL) {
		return false
	}
	return input.CalendarPlanID != nil || input.CalendarPlanNumber == nil
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
