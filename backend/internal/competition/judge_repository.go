package competition

import (
	"context"
	"time"
)

type CompetitionJudge struct {
	ID            int64
	CompetitionID int64
	JudgeUserID   int64
	FirstName     string
	LastName      string
	MiddleName    *string
	PositionID    int64
	PositionName  string
	AppointedAt   time.Time
}

type CreatedCompetitionJudge struct {
	ID            int64
	CompetitionID int64
	JudgeUserID   int64
	PositionID    int64
	AppointedAt   time.Time
}

type JudgeAssignmentValidation struct {
	CompetitionExists bool
	UserExists        bool
	JudgeRoleExists   bool
	PositionExists    bool
	AssignmentExists  bool
}

type JudgeRepository interface {
	GetCompetitionStatus(ctx context.Context, competitionID int64) (string, error)
	ListCompetitionJudges(ctx context.Context, competitionID int64) ([]CompetitionJudge, error)
	ValidateCompetitionJudgeAssignment(
		ctx context.Context, competitionID, judgeUserID, positionID int64,
	) (JudgeAssignmentValidation, error)
	AssignCompetitionJudge(
		ctx context.Context, competitionID, judgeUserID, positionID int64,
	) (CreatedCompetitionJudge, error)
	RemoveCompetitionJudge(ctx context.Context, competitionID, assignmentID int64) (bool, error)
}
