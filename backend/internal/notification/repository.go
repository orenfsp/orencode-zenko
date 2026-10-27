package notification

import (
	"context"

	"github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type Repository interface {
	GetUserNotifications(ctx context.Context, userID int64, limit, offset int32) ([]db.InappNotification, error)
	MarkAsRead(ctx context.Context, id, userID int64) error
}

type repository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &repository{queries: queries}
}

func (r *repository) GetUserNotifications(ctx context.Context, userID int64, limit, offset int32) ([]db.InappNotification, error) {
	return r.queries.GetUserNotifications(ctx, db.GetUserNotificationsParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *repository) MarkAsRead(ctx context.Context, id, userID int64) error {
	return r.queries.MarkNotificationAsRead(ctx, db.MarkNotificationAsReadParams{
		ID:     id,
		UserID: userID,
	})
}
