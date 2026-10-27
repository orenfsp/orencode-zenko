package rank

import (
	"context"
	"errors"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/rank/errs"
)

type repositoryStub struct {
	deleted int64
}

func (r *repositoryStub) CreateRank(context.Context, Create) (Response, error) {
	return Response{}, nil
}

func (r *repositoryStub) GetRank(context.Context, int32) (Response, error) {
	return Response{}, nil
}

func (r *repositoryStub) GetListRank(context.Context) ([]*Response, error) {
	return nil, nil
}

func (r *repositoryStub) DeleteExpired(context.Context) (int64, error) {
	return r.deleted, nil
}

func TestCleanupExpiredReturnsDeletedCount(t *testing.T) {
	service := NewService(&repositoryStub{deleted: 3})

	deleted, err := service.CleanupExpired(context.Background())
	if err != nil {
		t.Fatalf("CleanupExpired() error = %v", err)
	}
	if deleted != 3 {
		t.Fatalf("deleted = %d, want 3", deleted)
	}
}

func TestCreateRankRejectsNegativeTTL(t *testing.T) {
	service := NewService(&repositoryStub{})

	_, err := service.CreateRank(context.Background(), CreateDTO{Name: "Первый", TTL: -1})
	if !errors.Is(err, errs.ErrInvalidCreate) {
		t.Fatalf("CreateRank() error = %v, want ErrInvalidCreate", err)
	}
}
