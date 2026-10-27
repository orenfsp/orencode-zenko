package rank

import (
	"context"
	"strings"

	"github.com/orenfsp/orencode-zenko.git/internal/rank/errs"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateRank(ctx context.Context, dto CreateDTO) (ResponseDTO, error) {
	if strings.TrimSpace(dto.Name) == "" || dto.TTL < 0 {
		return ResponseDTO{}, errs.ErrInvalidCreate
	}
	rows, err := s.repo.CreateRank(ctx, Create{
		Name: dto.Name,
		TTL:  dto.TTL,
	})
	if err != nil {
		return ResponseDTO{}, err
	}
	return ResponseDTO{
		ID:        rows.ID,
		Name:      rows.Name,
		TTL:       rows.TTL,
		CreatedAt: rows.CreatedAt,
		UpdatedAt: rows.UpdatedAt,
	}, nil
}

func (s *Service) CleanupExpired(ctx context.Context) (int64, error) {
	return s.repo.DeleteExpired(ctx)
}

func (s *Service) GetRank(ctx context.Context, id int32) (ResponseDTO, error) {
	rows, err := s.repo.GetRank(ctx, id)
	if err != nil {
		return ResponseDTO{}, err
	}
	return ResponseDTO{
		ID:        rows.ID,
		Name:      rows.Name,
		TTL:       rows.TTL,
		CreatedAt: rows.CreatedAt,
		UpdatedAt: rows.UpdatedAt,
	}, nil
}

func (s *Service) GetRankList(ctx context.Context) ([]*ResponseDTO, error) {

	rows, err := s.repo.GetListRank(ctx)
	if err != nil {
		return nil, err
	}
	arry := make([]*ResponseDTO, 0, len(rows))
	for _, i := range rows {
		arry = append(arry, &ResponseDTO{
			ID:        i.ID,
			Name:      i.Name,
			TTL:       i.TTL,
			CreatedAt: i.CreatedAt,
			UpdatedAt: i.UpdatedAt,
		})
	}
	return arry, nil
}
