package team

import "context"

type Repository interface {
	Create(context.Context, string, []int64, []int64) (Squad, error)
	List(context.Context) ([]Squad, error)
	ListByCoach(context.Context, int64) ([]Squad, error)
}
