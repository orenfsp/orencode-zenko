package judgecategory

import "context"

type Repository interface {
	CategoryExists(ctx context.Context, categoryID int64) (bool, error)
	Create(ctx context.Context, judgeUserID int64, request CreateRequest) (Application, error)
	ListMine(ctx context.Context, judgeUserID int64) ([]Application, error)
	ListPending(ctx context.Context, limit, offset int32) ([]Application, error)
	Review(ctx context.Context, reviewerUserID, applicationID int64, input ReviewInput) (Application, error)
	GetCurrent(ctx context.Context, judgeUserID int64) (*CurrentCategory, error)
	ListAssignments(ctx context.Context, judgeUserID int64) ([]Assignment, error)
	ListCategories(ctx context.Context) ([]Category, error)
}

type RoleChecker interface {
	HasRole(ctx context.Context, userID int64, roleCode string) (bool, error)
}
