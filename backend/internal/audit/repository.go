package audit

import "context"

type Repository interface {
	Write(ctx context.Context, event EventCreate) error
	List(ctx context.Context, limit, offset int32) ([]Event, error)
}
