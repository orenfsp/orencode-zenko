package evsk

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidInput = errors.New("invalid evsk standard")

type Service struct{ repo Repository }

func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) Upsert(ctx context.Context, req UpsertRequest) (Standard, error) {
	if req.RankID <= 0 || req.DisciplineID <= 0 || (req.MinimumPlace == nil && req.MinimumScore == nil) {
		return Standard{}, ErrInvalidInput
	}
	if req.MinimumPlace != nil && *req.MinimumPlace <= 0 {
		return Standard{}, ErrInvalidInput
	}
	if req.MinimumScore != nil {
		if _, err := strconv.ParseFloat(strings.TrimSpace(*req.MinimumScore), 64); err != nil {
			return Standard{}, ErrInvalidInput
		}
	}
	return s.repo.Upsert(ctx, req)
}
func (s *Service) List(ctx context.Context) ([]Standard, error) { return s.repo.List(ctx) }
