package deadline

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct{ queries *db.Queries }

func NewRepository(queries *db.Queries) Repository { return &sqlRepository{queries: queries} }

func (r *sqlRepository) RoleCodes(ctx context.Context, userID int64) ([]string, error) {
	return r.queries.ListDeadlineRoleCodes(ctx, userID)
}

func (r *sqlRepository) Registration(ctx context.Context) ([]Item, error) {
	rows, err := r.queries.ListRegistrationDeadlines(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		id := row.ID
		items = append(items, Item{Type: "competition_registration", ID: row.ID, Title: row.Name, DueAt: row.DueAt.Time, Status: row.Status, CompetitionID: &id})
	}
	return items, nil
}

func (r *sqlRepository) AthleteRanks(ctx context.Context, userID int64) ([]Item, error) {
	rows, err := r.queries.ListAthleteRankDeadlines(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, credentialItem("rank_expiration", row.ID, row.UserID, row.LastName, row.FirstName, textValue(row.MiddleName), row.SubjectName, row.DueAt.Time))
	}
	return items, nil
}

func (r *sqlRepository) CoachAthleteRanks(ctx context.Context, userID int64) ([]Item, error) {
	rows, err := r.queries.ListCoachAthleteRankDeadlines(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, credentialItem("rank_expiration", row.ID, row.UserID, row.LastName, row.FirstName, textValue(row.MiddleName), row.SubjectName, row.DueAt.Time))
	}
	return items, nil
}

func (r *sqlRepository) AllRanks(ctx context.Context) ([]Item, error) {
	rows, err := r.queries.ListAllRankDeadlines(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, credentialItem("rank_expiration", row.ID, row.UserID, row.LastName, row.FirstName, textValue(row.MiddleName), row.SubjectName, row.DueAt.Time))
	}
	return items, nil
}

func (r *sqlRepository) JudgeCategories(ctx context.Context, userID int64) ([]Item, error) {
	rows, err := r.queries.ListJudgeCategoryDeadlines(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, credentialItem("judge_category_expiration", row.ID, row.UserID, row.LastName, row.FirstName, textValue(row.MiddleName), row.SubjectName, row.DueAt.Time))
	}
	return items, nil
}

func (r *sqlRepository) AllJudgeCategories(ctx context.Context) ([]Item, error) {
	rows, err := r.queries.ListAllJudgeCategoryDeadlines(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		items = append(items, credentialItem("judge_category_expiration", row.ID, row.UserID, row.LastName, row.FirstName, textValue(row.MiddleName), row.SubjectName, row.DueAt.Time))
	}
	return items, nil
}

func (r *sqlRepository) JudgeAssignments(ctx context.Context, userID int64) ([]Item, error) {
	rows, err := r.queries.ListJudgeAssignmentDeadlines(ctx, userID)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		id := row.CompetitionID
		items = append(items, Item{Type: "judge_assignment", ID: row.ID, Title: row.CompetitionName + " · " + row.PositionName, DueAt: row.DueAt.Time, Status: "appointed", CompetitionID: &id})
	}
	return items, nil
}

func (r *sqlRepository) AllJudgeAssignments(ctx context.Context) ([]Item, error) {
	rows, err := r.queries.ListAllJudgeAssignmentDeadlines(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		id := row.CompetitionID
		items = append(items, Item{
			Type: "judge_assignment", ID: row.ID, Title: row.CompetitionName + " · " + row.PositionName,
			DueAt: row.DueAt.Time, Status: "appointed", CompetitionID: &id,
		})
	}
	return items, nil
}

func (r *sqlRepository) ListReports(ctx context.Context) ([]Report, error) {
	rows, err := r.queries.ListReportDeadlines(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Report, 0, len(rows))
	for _, row := range rows {
		items = append(items, makeReport(row.ID, row.Name, row.DueAt, row.Status, row.DeadlineType, row.CreatedByUserID, row.CreatedAt, row.UpdatedAt))
	}
	return items, nil
}

func (r *sqlRepository) GetReport(ctx context.Context, id int64) (Report, error) {
	row, err := r.queries.GetReportDeadline(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Report{}, ErrNotFound
	}
	if err != nil {
		return Report{}, err
	}
	return makeReport(row.ID, row.Name, row.DueAt, row.Status, row.DeadlineType, row.CreatedByUserID, row.CreatedAt, row.UpdatedAt), nil
}

func (r *sqlRepository) CreateReport(ctx context.Context, actorID int64, name string, dueAt time.Time, status, reportType string) (Report, error) {
	row, err := r.queries.CreateReportDeadline(ctx, db.CreateReportDeadlineParams{Name: name, DueAt: pgtype.Timestamptz{Time: dueAt, Valid: true}, Status: status, DeadlineType: reportType, CreatedByUserID: actorID})
	if err != nil {
		return Report{}, err
	}
	return makeReport(row.ID, row.Name, row.DueAt, row.Status, row.DeadlineType, row.CreatedByUserID, row.CreatedAt, row.UpdatedAt), nil
}

func (r *sqlRepository) UpdateReport(ctx context.Context, report Report) (Report, error) {
	row, err := r.queries.UpdateReportDeadline(ctx, db.UpdateReportDeadlineParams{Name: report.Name, DueAt: pgtype.Timestamptz{Time: report.DueAt, Valid: true}, Status: report.Status, DeadlineType: report.Type, ID: report.ID})
	if errors.Is(err, pgx.ErrNoRows) {
		return Report{}, ErrNotFound
	}
	if err != nil {
		return Report{}, err
	}
	return makeReport(row.ID, row.Name, row.DueAt, row.Status, row.DeadlineType, row.CreatedByUserID, row.CreatedAt, row.UpdatedAt), nil
}

func (r *sqlRepository) CreateReportDeadlineReminders(ctx context.Context, daysBefore int32) error {
	return r.queries.CreateReportDeadlineReminders(ctx, daysBefore)
}

func makeReport(id int64, name string, dueAt pgtype.Timestamptz, status, reportType string, creator int64, created, updated pgtype.Timestamptz) Report {
	return Report{ID: id, Name: name, DueAt: dueAt.Time, Status: status, Type: reportType, CreatedByUserID: creator, CreatedAt: created.Time, UpdatedAt: updated.Time}
}

func credentialItem(kind string, id, userID int64, lastName, firstName, middleName, subject string, dueAt time.Time) Item {
	name := strings.Join(strings.Fields(strings.Join([]string{lastName, firstName, middleName}, " ")), " ")
	return Item{Type: kind, ID: id, Title: subject, DueAt: dueAt, Status: "active", UserID: &userID, UserName: &name}
}

func textValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
