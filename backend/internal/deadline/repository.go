package deadline

import (
	"context"
	"time"
)

type Repository interface {
	RoleCodes(ctx context.Context, userID int64) ([]string, error)
	Registration(ctx context.Context) ([]Item, error)
	AthleteRanks(ctx context.Context, userID int64) ([]Item, error)
	CoachAthleteRanks(ctx context.Context, coachUserID int64) ([]Item, error)
	AllRanks(ctx context.Context) ([]Item, error)
	JudgeCategories(ctx context.Context, userID int64) ([]Item, error)
	AllJudgeCategories(ctx context.Context) ([]Item, error)
	JudgeAssignments(ctx context.Context, userID int64) ([]Item, error)
	AllJudgeAssignments(ctx context.Context) ([]Item, error)
	ListReports(ctx context.Context) ([]Report, error)
	GetReport(ctx context.Context, id int64) (Report, error)
	CreateReport(ctx context.Context, actorID int64, name string, dueAt time.Time, status string) (Report, error)
	UpdateReport(ctx context.Context, report Report) (Report, error)
}

type PermissionChecker interface {
	HasPermission(ctx context.Context, userID int64, permissionCode string) (bool, error)
}
