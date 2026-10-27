package calendarplan

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) List(ctx context.Context, filter ListFilter) ([]CalendarPlan, error) {
	rows, err := r.queries.ListCalendarPlans(ctx, db.ListCalendarPlansParams{
		PYear: nullableInt16(filter.Year), POffset: filter.Offset, PLimit: filter.Limit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]CalendarPlan, 0, len(rows))
	for _, row := range rows {
		result = append(result, fromSQLCalendarPlan(row))
	}
	return result, nil
}

func (r *sqlRepository) Get(ctx context.Context, id int64) (CalendarPlan, error) {
	row, err := r.queries.GetCalendarPlan(ctx, id)
	if err != nil {
		return CalendarPlan{}, mapReadError(err)
	}
	return fromSQLCalendarPlan(row), nil
}

func (r *sqlRepository) Create(ctx context.Context, plan CalendarPlan) (CalendarPlan, error) {
	row, err := r.queries.CreateCalendarPlan(ctx, db.CreateCalendarPlanParams{
		PName: plan.Name, PYear: plan.Year,
	})
	if err != nil {
		return CalendarPlan{}, err
	}
	return fromSQLCalendarPlan(row), nil
}

func (r *sqlRepository) Update(ctx context.Context, plan UpdateCalendarPlan) (CalendarPlan, error) {
	row, err := r.queries.UpdateCalendarPlan(ctx, db.UpdateCalendarPlanParams{
		PName: plan.Name, PYear: plan.Year, PID: plan.ID,
	})
	if err != nil {
		return CalendarPlan{}, mapReadError(err)
	}
	return fromSQLCalendarPlan(row), nil
}

func (r *sqlRepository) ListCompetitions(
	ctx context.Context, planID int64, limit, offset int32,
) ([]Competition, error) {
	rows, err := r.queries.ListCalendarPlanCompetitions(ctx, db.ListCalendarPlanCompetitionsParams{
		PCalendarPlanID: planID, POffset: offset, PLimit: limit,
	})
	if err != nil {
		return nil, err
	}
	result := make([]Competition, 0, len(rows))
	for _, row := range rows {
		result = append(result, Competition{
			ID: row.ID, Name: row.Name, StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
			Status: row.Status, DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
			CalendarPlanNumber: textPointer(row.CalendarPlanNumber),
		})
	}
	return result, nil
}

func mapReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func fromSQLCalendarPlan(row db.CalendarPlan) CalendarPlan {
	return CalendarPlan{ID: row.ID, Name: row.Name, Year: row.Year}
}

func nullableInt16(value *int16) pgtype.Int2 {
	if value == nil {
		return pgtype.Int2{}
	}
	return pgtype.Int2{Int16: *value, Valid: true}
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	return &value.String
}
