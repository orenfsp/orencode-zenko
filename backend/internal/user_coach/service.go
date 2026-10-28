package usercoach

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func (s *Service) SearchParticipants(ctx context.Context, coachID int64, search string, disciplineID *int64, limit, offset int32) (ParticipantSearchResponse, error) {
	if err := validatePagination(limit, offset); err != nil {
		return ParticipantSearchResponse{}, err
	}
	if disciplineID != nil && *disciplineID <= 0 {
		return ParticipantSearchResponse{}, ErrInvalidInput
	}
	if err := s.requireRole(ctx, coachID, RoleCoach); err != nil {
		return ParticipantSearchResponse{}, err
	}
	items, total, err := s.repo.SearchParticipants(ctx, coachID, strings.TrimSpace(search), disciplineID, limit, offset)
	if err != nil {
		return ParticipantSearchResponse{}, err
	}
	if items == nil {
		items = []Participant{}
	}
	return ParticipantSearchResponse{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

const (
	RoleCoach       = "coach"
	RoleParticipant = "participant"

	DefaultLimit int32 = 20
	MaxLimit     int32 = 100
)

var (
	ErrInvalidInput      = errors.New("invalid user coach input")
	ErrInvalidPagination = errors.New("invalid pagination")
	ErrNotFound          = errors.New("user coach relation not found")
	ErrConflict          = errors.New("active relation already exists")
	ErrForbidden         = errors.New("operation forbidden")
)

type RoleChecker interface {
	HasRole(
		ctx context.Context,
		userID int64,
		role string,
	) (bool, error)
}

type Service struct {
	repo  Repository
	roles RoleChecker
}

func NewService(
	repo Repository,
	roles RoleChecker,
) (*Service, error) {
	if repo == nil || roles == nil {
		return nil, errors.New("user coach repository and role checker are required")
	}

	return &Service{
		repo:  repo,
		roles: roles,
	}, nil
}

func (s *Service) Create(
	ctx context.Context,
	coachUserID int64,
	req CreateRequest,
) (Response, error) {
	if coachUserID <= 0 ||
		req.ParticipantUserID <= 0 ||
		req.DisciplineID <= 0 ||
		coachUserID == req.ParticipantUserID {
		return Response{}, ErrInvalidInput
	}

	if err := s.requireRole(ctx, coachUserID, RoleCoach); err != nil {
		return Response{}, err
	}

	if err := s.requireRole(
		ctx,
		req.ParticipantUserID,
		RoleParticipant,
	); err != nil {
		return Response{}, err
	}

	relation, err := s.repo.Create(
		ctx,
		req.ParticipantUserID,
		coachUserID,
		req.DisciplineID,
	)
	if err != nil {
		return Response{}, err
	}

	return toResponse(relation), nil
}

func (s *Service) ListByCoach(
	ctx context.Context,
	coachUserID int64,
	limit int32,
	offset int32,
) ([]Response, error) {
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}

	if err := s.requireRole(ctx, coachUserID, RoleCoach); err != nil {
		return nil, err
	}

	relations, err := s.repo.ListByCoach(
		ctx,
		coachUserID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	return toResponses(relations), nil
}

func (s *Service) ListByParticipant(
	ctx context.Context,
	participantUserID int64,
	limit int32,
	offset int32,
) ([]Response, error) {
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}

	if err := s.requireRole(
		ctx,
		participantUserID,
		RoleParticipant,
	); err != nil {
		return nil, err
	}

	relations, err := s.repo.ListByParticipant(
		ctx,
		participantUserID,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	return toResponses(relations), nil
}

func (s *Service) Confirm(
	ctx context.Context,
	participantUserID int64,
	relationID int64,
) (Response, error) {
	if participantUserID <= 0 || relationID <= 0 {
		return Response{}, ErrInvalidInput
	}

	if err := s.requireRole(
		ctx,
		participantUserID,
		RoleParticipant,
	); err != nil {
		return Response{}, err
	}

	relation, err := s.repo.Confirm(
		ctx,
		relationID,
		participantUserID,
	)
	if err != nil {
		return Response{}, err
	}

	return toResponse(relation), nil
}

func (s *Service) End(
	ctx context.Context,
	actorUserID int64,
	relationID int64,
) (Response, error) {
	if actorUserID <= 0 || relationID <= 0 {
		return Response{}, ErrInvalidInput
	}

	isCoach, err := s.roles.HasRole(ctx, actorUserID, RoleCoach)
	if err != nil {
		return Response{}, fmt.Errorf("check coach role: %w", err)
	}

	isParticipant, err := s.roles.HasRole(
		ctx,
		actorUserID,
		RoleParticipant,
	)
	if err != nil {
		return Response{}, fmt.Errorf("check participant role: %w", err)
	}

	if !isCoach && !isParticipant {
		return Response{}, ErrForbidden
	}

	relation, err := s.repo.End(ctx, relationID, actorUserID)
	if err != nil {
		return Response{}, err
	}

	return toResponse(relation), nil
}

func (s *Service) requireRole(
	ctx context.Context,
	userID int64,
	role string,
) error {
	if userID <= 0 {
		return ErrForbidden
	}

	ok, err := s.roles.HasRole(ctx, userID, role)
	if err != nil {
		return fmt.Errorf("check user role: %w", err)
	}

	if !ok {
		return ErrForbidden
	}

	return nil
}

func validatePagination(limit, offset int32) error {
	if limit < 1 || limit > MaxLimit || offset < 0 {
		return ErrInvalidPagination
	}

	return nil
}

func toResponse(model UserCoach) Response {
	return Response{
		ID:                model.ID,
		ParticipantUserID: model.UserID,
		CoachUserID:       model.CoachUserID,
		DisciplineID:      model.DisciplineID,
		Status:            model.Status,
		RequestedAt:       model.RequestedAt,
		ConfirmedAt:       model.ConfirmedAt,
		EndedAt:           model.EndedAt,
	}
}

func toResponses(models []UserCoach) []Response {
	result := make([]Response, 0, len(models))

	for _, model := range models {
		result = append(result, toResponse(model))
	}

	return result
}
