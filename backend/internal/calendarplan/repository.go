package calendarplan

import (
	"context"
	"time"
)

type CalendarPlan struct {
	ID   int64
	Name string
	Year int16
}

type Competition struct {
	ID                 int64
	Name               string
	StartAt            time.Time
	EndAt              time.Time
	Status             string
	DisciplineID       int64
	DisciplineName     string
	CalendarPlanNumber *string
}

type ListFilter struct {
	Year   *int16
	Limit  int32
	Offset int32
}

type UpdateCalendarPlan struct {
	ID   int64
	Name string
	Year int16
}

type Repository interface {
	List(ctx context.Context, filter ListFilter) ([]CalendarPlan, error)
	Get(ctx context.Context, id int64) (CalendarPlan, error)
	Create(ctx context.Context, plan CalendarPlan) (CalendarPlan, error)
	Update(ctx context.Context, plan UpdateCalendarPlan) (CalendarPlan, error)
	ListCompetitions(ctx context.Context, planID int64, limit, offset int32) ([]Competition, error)
}
