package municipality

import "context"

type Repository interface {
	Create(ctx context.Context, input CreateMunicipality) (Municipality, error)
	GetByID(ctx context.Context, id int64) (Municipality, error)
	List(ctx context.Context, limit, offset int32) ([]Municipality, error)
}
