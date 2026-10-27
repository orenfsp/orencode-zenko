package judgecategory

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRepository(queries *db.Queries, pool *pgxpool.Pool) Repository {
	return &sqlRepository{queries: queries, pool: pool}
}

func (r *sqlRepository) CategoryExists(ctx context.Context, categoryID int64) (bool, error) {
	return r.queries.JudgeCategoryExists(ctx, categoryID)
}

func (r *sqlRepository) Create(ctx context.Context, judgeUserID int64, request CreateRequest) (Application, error) {
	row, err := r.queries.CreateJudgeCategoryApplication(ctx, db.CreateJudgeCategoryApplicationParams{
		JudgeUserID: judgeUserID, CategoryID: request.CategoryID, ApplicationType: request.ApplicationType,
	})
	if err != nil {
		return Application{}, mapWriteError(err)
	}
	return fromApplication(row), nil
}

func (r *sqlRepository) ListMine(ctx context.Context, judgeUserID int64) ([]Application, error) {
	rows, err := r.queries.ListJudgeCategoryApplicationsByJudge(ctx, judgeUserID)
	if err != nil {
		return nil, err
	}
	result := make([]Application, 0, len(rows))
	for _, row := range rows {
		categoryName := row.CategoryName
		result = append(result, Application{
			ID: row.ID, JudgeUserID: row.JudgeUserID, CategoryID: row.CategoryID,
			CategoryName: &categoryName, ApplicationType: row.ApplicationType, Status: row.Status,
			SubmittedAt: row.SubmittedAt.Time, ReviewedByUserID: intPointer(row.ReviewedByUserID),
			ReviewedAt: timePointer(row.ReviewedAt), AssignedAt: datePointer(row.AssignedAt),
			ExpiresAt: datePointer(row.ExpiresAt), OrderNumber: textPointer(row.OrderNumber),
			RejectionReason: textPointer(row.RejectionReason),
		})
	}
	return result, nil
}

func (r *sqlRepository) ListPending(ctx context.Context, limit, offset int32) ([]Application, error) {
	rows, err := r.queries.ListPendingJudgeCategoryApplications(ctx, db.ListPendingJudgeCategoryApplicationsParams{POffset: offset, PLimit: limit})
	if err != nil {
		return nil, err
	}
	result := make([]Application, 0, len(rows))
	for _, row := range rows {
		categoryName := row.CategoryName
		judgeName := fullName(row.LastName, row.FirstName, textValue(row.MiddleName))
		result = append(result, Application{
			ID: row.ID, JudgeUserID: row.JudgeUserID, JudgeName: &judgeName,
			CategoryID: row.CategoryID, CategoryName: &categoryName,
			ApplicationType: row.ApplicationType, Status: row.Status, SubmittedAt: row.SubmittedAt.Time,
		})
	}
	return result, nil
}

func (r *sqlRepository) Review(ctx context.Context, reviewerUserID, applicationID int64, input ReviewInput) (Application, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Application{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	current, err := queries.GetJudgeCategoryApplicationForUpdate(ctx, applicationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Application{}, ErrNotFound
	}
	if err != nil {
		return Application{}, err
	}
	if current.Status != "pending" {
		return Application{}, ErrConflict
	}

	updated, err := queries.ReviewJudgeCategoryApplication(ctx, db.ReviewJudgeCategoryApplicationParams{
		Status: input.Status, ReviewedByUserID: pgtype.Int8{Int64: reviewerUserID, Valid: true},
		AssignedAt: nullableDate(input.AssignedAt), ExpiresAt: nullableDate(input.ExpiresAt),
		OrderNumber: nullableText(input.OrderNumber), RejectionReason: nullableText(input.RejectionReason), ID: applicationID,
	})
	if err != nil {
		return Application{}, mapWriteError(err)
	}
	if input.Status == "approved" {
		_, err = queries.CreateUserJudgeCategoryFromApplication(ctx, db.CreateUserJudgeCategoryFromApplicationParams{
			UserID: current.JudgeUserID, CategoryID: current.CategoryID,
			AssignedAt: nullableDate(input.AssignedAt), ExpiresAt: nullableDate(input.ExpiresAt),
			OrderNumber: nullableText(input.OrderNumber),
		})
		if err != nil {
			return Application{}, mapWriteError(err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Application{}, err
	}
	return fromApplication(updated), nil
}

func (r *sqlRepository) GetCurrent(ctx context.Context, judgeUserID int64) (*CurrentCategory, error) {
	row, err := r.queries.GetCurrentJudgeCategory(ctx, judgeUserID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &CurrentCategory{
		ID: row.ID, UserID: row.UserID, CategoryID: row.CategoryID, CategoryName: row.CategoryName,
		AssignedAt: row.AssignedAt.Time, ExpiresAt: datePointer(row.ExpiresAt), OrderNumber: textPointer(row.OrderNumber),
	}, nil
}

func (r *sqlRepository) ListAssignments(ctx context.Context, judgeUserID int64) ([]Assignment, error) {
	rows, err := r.queries.ListJudgeCompetitionAssignments(ctx, judgeUserID)
	if err != nil {
		return nil, err
	}
	result := make([]Assignment, 0, len(rows))
	for _, row := range rows {
		result = append(result, Assignment{
			ID: row.ID, CompetitionID: row.CompetitionID, CompetitionName: row.CompetitionName,
			StartAt: row.StartAt.Time, EndAt: row.EndAt.Time, Status: row.Status,
			PositionID: row.PositionID, PositionName: row.PositionName, AppointedAt: row.AppointedAt.Time,
		})
	}
	return result, nil
}

func fromApplication(row db.JudgeCategoryApplication) Application {
	return Application{
		ID: row.ID, JudgeUserID: row.JudgeUserID, CategoryID: row.CategoryID,
		ApplicationType: row.ApplicationType, Status: row.Status, SubmittedAt: row.SubmittedAt.Time,
		ReviewedByUserID: intPointer(row.ReviewedByUserID), ReviewedAt: timePointer(row.ReviewedAt),
		AssignedAt: datePointer(row.AssignedAt), ExpiresAt: datePointer(row.ExpiresAt),
		OrderNumber: textPointer(row.OrderNumber), RejectionReason: textPointer(row.RejectionReason),
	}
}

func nullableDate(value *time.Time) pgtype.Date {
	if value == nil {
		return pgtype.Date{}
	}
	return pgtype.Date{Time: *value, Valid: true}
}

func nullableText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func intPointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func datePointer(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func textValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func fullName(lastName, firstName, middleName string) string {
	return strings.Join(strings.Fields(strings.Join([]string{lastName, firstName, middleName}, " ")), " ")
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	switch pgErr.Code {
	case "23505":
		return ErrConflict
	case "23503":
		return ErrCategoryMissing
	default:
		return err
	}
}
