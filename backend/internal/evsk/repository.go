package evsk

import "context"

type Repository interface {
	Upsert(context.Context, UpsertRequest) (Standard, error)
	List(context.Context) ([]Standard, error)
}
