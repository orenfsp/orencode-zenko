package competition

import "time"

type AssignJudgeRequest struct {
	JudgeUserID int64 `json:"judge_user_id"`
	PositionID  int64 `json:"position_id"`
}

type JudgeUserDTO struct {
	ID         int64   `json:"id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	MiddleName *string `json:"middle_name"`
}

type JudgePositionDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CompetitionJudgeDTO struct {
	ID            int64            `json:"id"`
	CompetitionID int64            `json:"competition_id"`
	Judge         JudgeUserDTO     `json:"judge"`
	Position      JudgePositionDTO `json:"position"`
	AppointedAt   time.Time        `json:"appointed_at"`
}

type CreatedCompetitionJudgeDTO struct {
	ID            int64     `json:"id"`
	CompetitionID int64     `json:"competition_id"`
	JudgeUserID   int64     `json:"judge_user_id"`
	PositionID    int64     `json:"position_id"`
	AppointedAt   time.Time `json:"appointed_at"`
}
type JudgeCompetitionDTO struct {
	ID              int64            `json:"id"`
	Name            string           `json:"name"`
	Discipline      ReferenceDTO     `json:"discipline"`
	CompetitionType ReferenceDTO     `json:"competition_type"`
	Format          string           `json:"format"`
	StartAt         time.Time        `json:"start_at"`
	EndAt           time.Time        `json:"end_at"`
	Status          string           `json:"status"`
	Position        JudgePositionDTO `json:"position"`
	AppointedAt     time.Time        `json:"appointed_at"`
}
