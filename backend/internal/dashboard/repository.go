package dashboard

import "context"

type Repository interface {
	Load(ctx context.Context, limit, days int32) (Data, error)
}
