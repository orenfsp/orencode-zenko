package team

import (
	"context"
	"errors"
	usercoach "github.com/orenfsp/orencode-zenko.git/internal/user_coach"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid squad input")
	ErrForbidden    = errors.New("squad operation forbidden")
)

type RoleChecker interface {
	HasRole(context.Context, int64, string) (bool, error)
}
type Relations interface {
	ListByCoach(context.Context, int64, int32, int32) ([]usercoach.UserCoach, error)
}
type Service struct {
	repository Repository
	roles      RoleChecker
	relations  Relations
}

func NewService(repository Repository, roles RoleChecker, relations Relations) *Service {
	return &Service{repository: repository, roles: roles, relations: relations}
}
func (s *Service) Create(ctx context.Context, actorID int64, request CreateRequest) (Squad, error) {
	name := strings.TrimSpace(request.Name)
	if actorID <= 0 || name == "" || len(name) > 200 || len(request.MemberIDs) == 0 {
		return Squad{}, ErrInvalidInput
	}
	memberIDs, coachIDs := unique(request.MemberIDs), unique(request.CoachIDs)
	isAdmin, err := s.roles.HasRole(ctx, actorID, "admin")
	if err != nil {
		return Squad{}, err
	}
	if !isAdmin {
		isAdmin, err = s.roles.HasRole(ctx, actorID, "super_admin")
		if err != nil {
			return Squad{}, err
		}
	}
	if !isAdmin {
		isCoach, err := s.roles.HasRole(ctx, actorID, "coach")
		if err != nil {
			return Squad{}, err
		}
		if !isCoach || !contains(coachIDs, actorID) {
			return Squad{}, ErrForbidden
		}
		if s.relations == nil {
			return Squad{}, ErrForbidden
		}
		relations, err := s.relations.ListByCoach(ctx, actorID, 100, 0)
		if err != nil {
			return Squad{}, err
		}
		confirmed := map[int64]bool{}
		for _, relation := range relations {
			if relation.Status == usercoach.StatusConfirmed {
				confirmed[relation.UserID] = true
			}
		}
		for _, memberID := range memberIDs {
			if !confirmed[memberID] {
				return Squad{}, ErrForbidden
			}
		}
	}
	return s.repository.Create(ctx, name, memberIDs, coachIDs)
}
func (s *Service) List(ctx context.Context, actorID int64) ([]Squad, error) {
	isAdmin, err := s.roles.HasRole(ctx, actorID, "admin")
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		isAdmin, err = s.roles.HasRole(ctx, actorID, "super_admin")
		if err != nil {
			return nil, err
		}
	}
	if isAdmin {
		return s.repository.List(ctx)
	}
	return s.repository.ListByCoach(ctx, actorID)
}
func unique(values []int64) []int64 {
	seen := map[int64]bool{}
	result := make([]int64, 0, len(values))
	for _, v := range values {
		if v > 0 && !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

func contains(values []int64, target int64) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
