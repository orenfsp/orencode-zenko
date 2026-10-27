package rank

import (
	"context"
)

type Repository interface {
	CreateRank(ctx context.Context, createData Create) (Response, error) 
	GetRank(ctx context.Context, id int32) (Response, error)
	GetListRank(ctx context.Context) ([]*Response, error)
}