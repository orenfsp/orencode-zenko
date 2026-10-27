package municipality

import (
	"context"
	"strings"
	"unicode/utf8"

	municipalityerrs "github.com/orenfsp/orencode-zenko.git/internal/municipality/errs"
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

func (s *Service) Create(ctx context.Context, req CreateRequest) (ResponseDTO, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" || utf8.RuneCountInString(name) > 200 {
		return ResponseDTO{}, municipalityerrs.ErrInvalidCreate
	}
	created, err := s.repo.Create(ctx, CreateMunicipality{Name: name})
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(created), nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (ResponseDTO, error) {
	if id <= 0 {
		return ResponseDTO{}, municipalityerrs.ErrInvalidID
	}
	item, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ResponseDTO{}, err
	}
	return toResponse(item), nil
}

func (s *Service) List(ctx context.Context, limit, offset int32) ([]ResponseDTO, error) {
	if limit < 1 || limit > MaxLimit || offset < 0 {
		return nil, municipalityerrs.ErrInvalidPagination
	}
	items, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]ResponseDTO, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result, nil
}

func toResponse(item Municipality) ResponseDTO {
	return ResponseDTO{ID: item.ID, Name: item.Name}
}
