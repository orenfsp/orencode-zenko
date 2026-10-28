package dashboard

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlRepository struct {
	queries *db.Queries
}

func NewRepository(queries *db.Queries) Repository {
	return &sqlRepository{queries: queries}
}

func (r *sqlRepository) Load(ctx context.Context, limit, days int32, year int) (Data, error) {
	periodStart := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(1, 0, 0)
	metrics, err := r.queries.GetFederationDashboardMetrics(ctx, db.GetFederationDashboardMetricsParams{
		PeriodStart: pgtype.Timestamptz{Time: periodStart, Valid: true},
		PeriodEnd:   pgtype.Timestamptz{Time: periodEnd, Valid: true},
	})
	if err != nil {
		return Data{}, err
	}

	competitions, err := r.queries.ListDashboardUpcomingCompetitions(ctx, limit)
	if err != nil {
		return Data{}, err
	}
	competitionApplications, err := r.queries.ListDashboardPendingCompetitionApplications(ctx, limit)
	if err != nil {
		return Data{}, err
	}
	rankApplications, err := r.queries.ListDashboardPendingRankApplications(ctx, limit)
	if err != nil {
		return Data{}, err
	}
	judgeApplications, err := r.queries.ListDashboardPendingJudgeCategoryApplications(ctx, limit)
	if err != nil {
		return Data{}, err
	}
	expiringRanks, err := r.queries.ListDashboardExpiringRanks(ctx, db.ListDashboardExpiringRanksParams{PDays: days, PLimit: limit})
	if err != nil {
		return Data{}, err
	}
	expiringJudges, err := r.queries.ListDashboardExpiringJudgeCategories(ctx, db.ListDashboardExpiringJudgeCategoriesParams{PDays: days, PLimit: limit})
	if err != nil {
		return Data{}, err
	}
	deadlines, err := r.queries.ListDashboardRegistrationDeadlines(ctx, db.ListDashboardRegistrationDeadlinesParams{PDays: days, PLimit: limit})
	if err != nil {
		return Data{}, err
	}
	reportDeadlines, err := r.queries.ListDashboardReportDeadlines(ctx, limit)
	if err != nil {
		return Data{}, err
	}

	data := emptyData()
	data.PeriodYear = year
	data.Metrics = Metrics{
		Athletes: metrics.AthleteCount, Competitions: metrics.CompetitionCount,
		Participants: metrics.ParticipantCount, ActiveRanks: metrics.ActiveRankCount,
	}
	for _, row := range competitions {
		data.UpcomingCompetitions = append(data.UpcomingCompetitions, UpcomingCompetition{
			ID: row.ID, Name: row.Name, StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
			Status: row.Status, DisciplineName: row.DisciplineName,
		})
	}
	for _, row := range competitionApplications {
		competitionID, competitionName := row.CompetitionID, row.CompetitionName
		data.PendingCompetitionApplications = append(data.PendingCompetitionApplications, PendingApplication{
			ID: row.ID, SubmittedAt: row.SubmittedAt.Time, AthleteUserID: row.AthleteUserID,
			AthleteName:   fullName(row.LastName, row.FirstName, nullableText(row.MiddleName)),
			CompetitionID: &competitionID, CompetitionName: &competitionName,
		})
	}
	for _, row := range rankApplications {
		rankName, disciplineName := row.RankName, row.DisciplineName
		data.PendingRankApplications = append(data.PendingRankApplications, PendingApplication{
			ID: row.ID, SubmittedAt: row.CreatedAt.Time, AthleteUserID: row.AthleteUserID,
			AthleteName: fullName(row.LastName, row.FirstName, nullableText(row.MiddleName)),
			RankName:    &rankName, DisciplineName: &disciplineName,
		})
	}
	for _, row := range judgeApplications {
		categoryName, applicationType := row.CategoryName, row.ApplicationType
		data.PendingJudgeApplications = append(data.PendingJudgeApplications, PendingApplication{
			ID: row.ID, SubmittedAt: row.SubmittedAt.Time, AthleteUserID: row.JudgeUserID,
			AthleteName:  fullName(row.LastName, row.FirstName, nullableText(row.MiddleName)),
			CategoryName: &categoryName, ApplicationType: &applicationType,
		})
	}
	for _, row := range expiringRanks {
		disciplineName := row.DisciplineName
		data.ExpiringRanks = append(data.ExpiringRanks, ExpiringCredential{
			ID: row.ID, UserID: row.UserID,
			UserName:       fullName(row.LastName, row.FirstName, nullableText(row.MiddleName)),
			CredentialName: row.CredentialName, DisciplineName: &disciplineName, ExpiresAt: row.ExpiresAt.Time,
		})
	}
	for _, row := range expiringJudges {
		data.ExpiringJudgeCategories = append(data.ExpiringJudgeCategories, ExpiringCredential{
			ID: row.ID, UserID: row.UserID,
			UserName:       fullName(row.LastName, row.FirstName, nullableText(row.MiddleName)),
			CredentialName: row.CredentialName, ExpiresAt: row.ExpiresAt.Time,
		})
	}
	for _, row := range deadlines {
		competitionID := row.CompetitionID
		data.Deadlines = append(data.Deadlines, Deadline{
			Type: "competition_registration", Title: row.CompetitionName,
			DueAt: row.DueAt.Time, CompetitionID: &competitionID,
		})
	}
	for _, row := range reportDeadlines {
		data.Deadlines = append(data.Deadlines, Deadline{
			Type: "report", Title: row.Name, DueAt: row.DueAt.Time,
		})
	}
	return data, nil
}

func emptyData() Data {
	return Data{
		UpcomingCompetitions:           make([]UpcomingCompetition, 0),
		PendingCompetitionApplications: make([]PendingApplication, 0),
		PendingRankApplications:        make([]PendingApplication, 0),
		PendingJudgeApplications:       make([]PendingApplication, 0),
		ExpiringRanks:                  make([]ExpiringCredential, 0),
		ExpiringJudgeCategories:        make([]ExpiringCredential, 0),
		Deadlines:                      make([]Deadline, 0),
	}
}

func nullableText(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func fullName(lastName, firstName, middleName string) string {
	return strings.Join(strings.Fields(strings.Join([]string{lastName, firstName, middleName}, " ")), " ")
}
