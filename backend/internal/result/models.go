package result

import (
	"io"
	"time"
)

const (
	ModeIndividual = "individual"
	ModeTeam       = "team"

	SheetDraft    = "draft"
	SheetApproved = "approved"

	ResultFinished     = "finished"
	ResultDidNotStart  = "did_not_start"
	ResultDisqualified = "disqualified"

	TaskNotAttempted = "not_attempted"
	TaskAttempted    = "attempted"
	TaskSolved       = "solved"

	AlgorithmicDiscipline = "программирование алгоритмическое"
	MaxTaskCount          = 50
	ScoringAlgorithmic    = "algorithmic"
	ScoringManual         = "manual"
	MaxManualTextLength   = 2000
)

type EventContext struct {
	EventID           int64
	EventName         string
	CompetitionID     int64
	CompetitionName   string
	CompetitionStatus string
	StartAt           time.Time
	EndAt             time.Time
	DisciplineID      int64
	DisciplineName    string
	AgeGroupName      *string
}

type Application struct {
	ID            int64
	AthleteUserID int64
	LastName      string
	FirstName     string
	MiddleName    *string
}

type TeamMembership struct {
	ApplicationID int64
	AthleteUserID int64
	TeamID        int64
	TeamName      string
}

type TeamSetup struct {
	TeamID         int64
	ApplicationIDs []int64
}

type Sheet struct {
	ID                 int64
	EventID            int64
	Mode               string
	TaskCount          int16
	Status             string
	CalculatedAt       *time.Time
	ApprovedByUserID   *int64
	ApprovedAt         *time.Time
	ProtocolStorageKey *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Revision           int64
	ScoringMode        string
	TemplateID         *int64
	TemplateSnapshot   *TemplateSnapshot
}

type TaskResult struct {
	TaskNumber           int16
	Status               string
	SolveTimeDeciseconds *int64
	NegativeAttempts     int32
}

type Entry struct {
	ID                       int64
	TeamID                   *int64
	TeamName                 *string
	Applications             []Application
	ResultStatus             string
	Tasks                    []TaskResult
	SolvedCount              int32
	TotalTimeDeciseconds     int64
	LastSolveTimeDeciseconds int64
	NegativeAttempts         int32
	Place                    *int32
	CustomValues             map[string]any
}

type Table struct {
	Context EventContext
	Sheet   Sheet
	Entries []Entry
	Judges  []Judge
}

type Judge struct {
	PositionName string
	LastName     string
	FirstName    string
	MiddleName   *string
}

type EntryUpdate struct {
	EntryID      int64
	ResultStatus string
	Tasks        []TaskResult
}

type ManualEntryUpdate struct {
	EntryID      int64
	ResultStatus string
	Place        *int32
	CustomValues map[string]any
}

type EntrySummary struct {
	EntryID                  int64
	SolvedCount              int32
	TotalTimeDeciseconds     int64
	LastSolveTimeDeciseconds int64
	NegativeAttempts         int32
	Place                    *int32
}

type Approval struct {
	ApprovedAt  time.Time
	ProtocolURL string
}

type ProtocolDownload struct {
	Content  io.ReadCloser
	Filename string
}

type TemplateSnapshot struct {
	ID     int64           `json:"id"`
	Name   string          `json:"name"`
	Fields []TemplateField `json:"fields"`
}

type ResultTemplate struct {
	ID           int64           `json:"id"`
	DisciplineID int64           `json:"discipline_id"`
	Name         string          `json:"name"`
	Fields       []TemplateField `json:"fields"`
	IsActive     bool            `json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}
