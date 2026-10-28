package rating

import (
	"context"
	"time"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, filter Filter) (Page, error) {
	if filter.Year < 2000 || filter.Year > int32(time.Now().Year()+1) || filter.Limit < 1 || filter.Limit > 100 || filter.Offset < 0 {
		return Page{}, ErrInvalidInput
	}
	items, err := s.repository.List(ctx, filter)
	if err != nil {
		return Page{}, err
	}
	if items == nil {
		items = make([]Entry, 0)
	}
	return Page{Items: items, Limit: filter.Limit, Offset: filter.Offset}, nil
}
