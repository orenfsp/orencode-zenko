package competition

import (
	"context"

	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
)

const JudgeRoleCode = "judge"

type JudgeService struct {
	repository JudgeRepository
}

func NewJudgeService(repository JudgeRepository) *JudgeService {
	return &JudgeService{repository: repository}
}

func (s *JudgeService) ListPublic(ctx context.Context, competitionID int64) ([]CompetitionJudgeDTO, error) {
	if competitionID <= 0 {
		return nil, ErrInvalidJudgeRequest
	}
	status, err := s.repository.GetCompetitionStatus(ctx, competitionID)
	if err != nil {
		return nil, err
	}
	if !isPublicCompetitionStatus(status) {
		return nil, ErrJudgeCompetitionPrivate
	}
	items, err := s.repository.ListCompetitionJudges(ctx, competitionID)
	if err != nil {
		return nil, err
	}
	result := make([]CompetitionJudgeDTO, 0, len(items))
	for _, item := range items {
		result = append(result, CompetitionJudgeDTO{
			ID: item.ID, CompetitionID: item.CompetitionID, AppointedAt: item.AppointedAt,
			Judge: JudgeUserDTO{
				ID: item.JudgeUserID, FirstName: item.FirstName,
				LastName: item.LastName, MiddleName: item.MiddleName,
			},
			Position: JudgePositionDTO{ID: item.PositionID, Name: item.PositionName},
		})
	}
	return result, nil
}

func (s *JudgeService) Assign(
	ctx context.Context, competitionID int64, request AssignJudgeRequest,
) (CreatedCompetitionJudgeDTO, error) {
	if competitionID <= 0 || request.JudgeUserID <= 0 || request.PositionID <= 0 {
		return CreatedCompetitionJudgeDTO{}, ErrInvalidJudgeRequest
	}
	if err := s.requireDraft(ctx, competitionID); err != nil {
		return CreatedCompetitionJudgeDTO{}, err
	}
	validation, err := s.repository.ValidateCompetitionJudgeAssignment(
		ctx, competitionID, request.JudgeUserID, request.PositionID,
	)
	if err != nil {
		return CreatedCompetitionJudgeDTO{}, err
	}
	switch {
	case !validation.CompetitionExists:
		return CreatedCompetitionJudgeDTO{}, ErrJudgeCompetitionMissing
	case !validation.UserExists:
		return CreatedCompetitionJudgeDTO{}, ErrJudgeUserMissing
	case !validation.JudgeRoleExists:
		return CreatedCompetitionJudgeDTO{}, ErrUserIsNotJudge
	case !validation.PositionExists:
		return CreatedCompetitionJudgeDTO{}, ErrJudgePositionMissing
	case validation.AssignmentExists:
		return CreatedCompetitionJudgeDTO{}, ErrJudgeAssignmentExists
	}
	created, err := s.repository.AssignCompetitionJudge(
		ctx, competitionID, request.JudgeUserID, request.PositionID,
	)
	if err != nil {
		return CreatedCompetitionJudgeDTO{}, err
	}
	return CreatedCompetitionJudgeDTO{
		ID: created.ID, CompetitionID: created.CompetitionID,
		JudgeUserID: created.JudgeUserID, PositionID: created.PositionID,
		AppointedAt: created.AppointedAt,
	}, nil
}

func (s *JudgeService) Remove(ctx context.Context, competitionID, assignmentID int64) error {
	if competitionID <= 0 || assignmentID <= 0 {
		return ErrInvalidJudgeRequest
	}
	if err := s.requireDraft(ctx, competitionID); err != nil {
		return err
	}
	removed, err := s.repository.RemoveCompetitionJudge(ctx, competitionID, assignmentID)
	if err != nil {
		return err
	}
	if !removed {
		if err := s.requireDraft(ctx, competitionID); err != nil {
			return err
		}
		return ErrJudgeAssignmentMissing
	}
	return nil
}

func (s *JudgeService) requireDraft(ctx context.Context, competitionID int64) error {
	status, err := s.repository.GetCompetitionStatus(ctx, competitionID)
	if err != nil {
		return err
	}
	if status != "draft" {
		return ErrJudgeCompetitionLocked
	}
	return nil
}
func (s *JudgeService) ListByJudgeUser(
	ctx context.Context, judgeUserID int64, limit, offset int32,
) ([]JudgeCompetitionDTO, error) {
	if judgeUserID <= 0 {
		return nil, ErrInvalidJudgeRequest
	}
	if limit < 1 || limit > MaxLimit || offset < 0 {
		return nil, competitionerrs.ErrInvalidPagination
	}
	items, err := s.repository.ListCompetitionsByJudgeUser(ctx, judgeUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	result := make([]JudgeCompetitionDTO, 0, len(items))
	for _, item := range items {
		result = append(result, JudgeCompetitionDTO{
			ID: item.ID, Name: item.Name,
			Discipline:      ReferenceDTO{ID: item.DisciplineID, Name: item.DisciplineName},
			CompetitionType: ReferenceDTO{ID: item.CompetitionTypeID, Name: item.CompetitionTypeName},
			Format:          item.Format, StartAt: item.StartAt, EndAt: item.EndAt, Status: item.Status,
			Position:    JudgePositionDTO{ID: item.PositionID, Name: item.PositionName},
			AppointedAt: item.AppointedAt,
		})
	}
	return result, nil
}
