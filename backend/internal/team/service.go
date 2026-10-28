package team

import (
	"context"
	"errors"
	"strings"
)

var ErrInvalidInput = errors.New("invalid squad input")

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }
func (s *Service) Create(ctx context.Context, request CreateRequest) (Squad, error) {
	name := strings.TrimSpace(request.Name)
	if name == "" || len(name) > 200 || len(request.MemberIDs) == 0 {
		return Squad{}, ErrInvalidInput
	}
	return s.repository.Create(ctx, name, unique(request.MemberIDs), unique(request.CoachIDs))
}
func (s *Service) List(ctx context.Context) ([]Squad, error) { return s.repository.List(ctx) }
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
