package registry

import (
	"context"
	"strings"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) ListUsers(ctx context.Context, filter UserFilter) (Page[User], error) {
	filter.Search = normalizedString(filter.Search)
	filter.RoleCode = normalizedString(filter.RoleCode)
	if !validPagination(filter.Limit, filter.Offset) || !validIDs(
		filter.DisciplineID, filter.OrganizationID, filter.RankID,
		filter.AgeGroupID, filter.MunicipalityID,
	) {
		return Page[User]{}, ErrInvalidFilter
	}
	return s.repository.ListUsers(ctx, filter)
}

func (s *Service) ListOrganizations(ctx context.Context, filter OrganizationFilter) (Page[Organization], error) {
	filter.Search = normalizedString(filter.Search)
	if !validPagination(filter.Limit, filter.Offset) || !validIDs(filter.MunicipalityID) {
		return Page[Organization]{}, ErrInvalidFilter
	}
	return s.repository.ListOrganizations(ctx, filter)
}

func normalizedString(value *string) *string {
	if value == nil {
		return nil
	}
	normalized := strings.TrimSpace(*value)
	if normalized == "" {
		return nil
	}
	return &normalized
}

func validPagination(limit, offset int32) bool { return limit >= 1 && limit <= MaxLimit && offset >= 0 }

func validIDs(values ...*int64) bool {
	for _, value := range values {
		if value != nil && *value <= 0 {
			return false
		}
	}
	return true
}
