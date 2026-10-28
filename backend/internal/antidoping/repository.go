package antidoping

import (
	"context"
	"errors"
)

var (
	ErrInvalidInput = errors.New("invalid anti-doping certificate input")
	ErrNotFound     = errors.New("anti-doping certificate not found")
	ErrConflict     = errors.New("anti-doping certificate was already reviewed")
)

type Repository interface {
	Create(ctx context.Context, userID, fileID int64) (Certificate, error)
	ListByUser(ctx context.Context, userID int64) ([]Certificate, error)
	Review(ctx context.Context, reviewerUserID, certificateID int64, input ReviewInput) (Certificate, error)
	Expire(ctx context.Context) (int64, error)
}

type AllRepository interface {
	Repository
	ListAll(ctx context.Context) ([]Certificate, error)
}
