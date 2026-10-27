package rbac

import (
	"context"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"
)

var roleCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{2,99}$`)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateRole(ctx context.Context, request CreateRoleRequest) (Role, error) {
	request.Code = strings.ToLower(strings.TrimSpace(request.Code))
	request.Name = strings.TrimSpace(request.Name)
	if !roleCodePattern.MatchString(request.Code) || request.Name == "" || utf8.RuneCountInString(request.Name) > 100 {
		return Role{}, ErrInvalidRole
	}
	request.PermissionCodes = uniqueStrings(request.PermissionCodes)
	return s.repository.CreateRole(ctx, request)
}

func (s *Service) AssignRole(ctx context.Context, userID, roleID int64) error {
	if userID <= 0 || roleID <= 0 {
		return ErrInvalidRole
	}
	return s.repository.AssignRole(ctx, userID, roleID)
}

func (s *Service) ListRoles(ctx context.Context) ([]Role, error) {
	return s.repository.ListRoles(ctx)
}

func (s *Service) ListPermissions(ctx context.Context) ([]Permission, error) {
	return s.repository.ListPermissions(ctx)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func (s *Service) ListSelfAssignableRoles(ctx context.Context) ([]Role, error) {
	return s.repository.ListSelfAssignableRoles(ctx)
}
