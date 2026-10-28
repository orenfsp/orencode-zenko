package rating

import "context"

type Repository interface {
	List(ctx context.Context, filter Filter) ([]Entry, error)
}
