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

type JudgeCompetition struct {
	ID                  int64
	Name                string
	DisciplineID        int64
	DisciplineName      string
	CompetitionTypeID   int64
	CompetitionTypeName string
	Format              string
	StartAt             time.Time
	EndAt               time.Time
	Status              string
	PositionID          int64
	PositionName        string
	AppointedAt         time.Time
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

type JudgeAssignment struct {
	AssignmentID        int64
	PositionID          int64
	PositionName        string
	AppointedAt         time.Time
	CompetitionID       int64
	CompetitionName     string
	DisciplineID        int64
	DisciplineName      string
	CompetitionTypeID   int64
	CompetitionTypeName string
	Format              string
	StartAt             time.Time
	EndAt               time.Time
	Status              string
}

type JudgeRepository interface {
	GetCompetitionStatus(ctx context.Context, competitionID int64) (string, error)
	ListCompetitionJudges(ctx context.Context, competitionID int64) ([]CompetitionJudge, error)
	ListAdminCompetitionJudges(ctx context.Context, competitionID int64) ([]CompetitionJudge, error)
	ValidateCompetitionJudgeAssignment(
		ctx context.Context, competitionID, judgeUserID, positionID int64,
	) (JudgeAssignmentValidation, error)
	AssignCompetitionJudge(
		ctx context.Context, competitionID, judgeUserID, positionID int64,
	) (CreatedCompetitionJudge, error)
	RemoveCompetitionJudge(ctx context.Context, competitionID, assignmentID int64) (bool, error)
	ListCompetitionsByJudgeUser(ctx context.Context, judgeUserID int64, limit, offset int32) ([]JudgeCompetition, error)
	ListJudgeAssignments(
		ctx context.Context, judgeUserID int64, limit, offset int32,
	) ([]JudgeAssignment, error)
}
