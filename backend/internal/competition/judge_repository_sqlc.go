package competition

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type sqlJudgeRepository struct {
	queries *db.Queries
}

func NewJudgeRepository(queries *db.Queries) JudgeRepository {
	return &sqlJudgeRepository{queries: queries}
}

func (r *sqlJudgeRepository) GetCompetitionStatus(ctx context.Context, competitionID int64) (string, error) {
	status, err := r.queries.GetCompetitionStatus(ctx, competitionID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrJudgeCompetitionMissing
	}
	return status, err
}

func (r *sqlJudgeRepository) ListCompetitionJudges(
	ctx context.Context, competitionID int64,
) ([]CompetitionJudge, error) {
	rows, err := r.queries.ListCompetitionJudges(ctx, db.ListCompetitionJudgesParams{
		PCompetitionID: competitionID, PPublicStatuses: publicCompetitionStatuses,
	})
	if err != nil {
		return nil, err
	}
	result := make([]CompetitionJudge, 0, len(rows))
	for _, row := range rows {
		result = append(result, CompetitionJudge{
			ID: row.ID, CompetitionID: row.CompetitionID, JudgeUserID: row.JudgeUserID,
			FirstName: row.FirstName, LastName: row.LastName, MiddleName: textPointer(row.MiddleName),
			PositionID: row.PositionID, PositionName: row.PositionName, AppointedAt: row.AppointedAt.Time,
		})
	}
	return result, nil
}

func (r *sqlJudgeRepository) ListAdminCompetitionJudges(
	ctx context.Context, competitionID int64,
) ([]CompetitionJudge, error) {
	rows, err := r.queries.ListAdminCompetitionJudges(ctx, competitionID)
	if err != nil {
		return nil, err
	}
	result := make([]CompetitionJudge, 0, len(rows))
	for _, row := range rows {
		result = append(result, CompetitionJudge{
			ID: row.ID, CompetitionID: row.CompetitionID, JudgeUserID: row.JudgeUserID,
			FirstName: row.FirstName, LastName: row.LastName, MiddleName: textPointer(row.MiddleName),
			PositionID: row.PositionID, PositionName: row.PositionName, AppointedAt: row.AppointedAt.Time,
		})
	}
	return result, nil
}

func (r *sqlJudgeRepository) ValidateCompetitionJudgeAssignment(
	ctx context.Context, competitionID, judgeUserID, positionID int64,
) (JudgeAssignmentValidation, error) {
	row, err := r.queries.ValidateCompetitionJudgeAssignment(ctx, db.ValidateCompetitionJudgeAssignmentParams{
		PCompetitionID: competitionID, PJudgeUserID: judgeUserID,
		PJudgeRoleCode: JudgeRoleCode, PPositionID: positionID,
	})
	if err != nil {
		return JudgeAssignmentValidation{}, err
	}
	return JudgeAssignmentValidation{
		CompetitionExists: row.CompetitionExists, UserExists: row.UserExists,
		JudgeRoleExists: row.JudgeRoleExists, PositionExists: row.PositionExists,
		AssignmentExists: row.AssignmentExists,
	}, nil
}

func (r *sqlJudgeRepository) AssignCompetitionJudge(
	ctx context.Context, competitionID, judgeUserID, positionID int64,
) (CreatedCompetitionJudge, error) {
	row, err := r.queries.AssignCompetitionJudge(ctx, db.AssignCompetitionJudgeParams{
		PCompetitionID: competitionID, PJudgeUserID: judgeUserID, PPositionID: positionID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CreatedCompetitionJudge{}, r.classifyAssignNoRows(
				ctx, competitionID, judgeUserID, positionID,
			)
		}
		return CreatedCompetitionJudge{}, mapJudgeWriteError(err)
	}
	return CreatedCompetitionJudge{
		ID: row.ID, CompetitionID: row.CompetitionID, JudgeUserID: row.JudgeUserID,
		PositionID: row.PositionID, AppointedAt: row.AppointedAt.Time,
	}, nil
}

func (r *sqlJudgeRepository) RemoveCompetitionJudge(
	ctx context.Context, competitionID, assignmentID int64,
) (bool, error) {
	rowsAffected, err := r.queries.RemoveCompetitionJudge(ctx, db.RemoveCompetitionJudgeParams{
		PCompetitionID: competitionID, PAssignmentID: assignmentID,
	})
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (r *sqlJudgeRepository) classifyAssignNoRows(
	ctx context.Context, competitionID, judgeUserID, positionID int64,
) error {
	status, err := r.GetCompetitionStatus(ctx, competitionID)
	if err != nil {
		return err
	}
	if status != "draft" {
		return ErrJudgeCompetitionLocked
	}
	validation, err := r.ValidateCompetitionJudgeAssignment(ctx, competitionID, judgeUserID, positionID)
	if err != nil {
		return err
	}
	if validation.AssignmentExists {
		return ErrJudgeAssignmentExists
	}
	return pgx.ErrNoRows
}

func mapJudgeWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrInvalidJudgeReference
	}
	return err
}
