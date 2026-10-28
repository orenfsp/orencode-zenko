package result

import "time"

type SettingsRequest struct {
	Mode         string             `json:"mode"`
	TaskCount    int16              `json:"task_count"`
	Teams        []TeamSetupRequest `json:"teams,omitempty"`
	ConfirmReset bool               `json:"confirm_reset,omitempty"`
	ScoringMode  string             `json:"scoring_mode,omitempty"`
	TemplateID   *int64             `json:"template_id,omitempty"`
}

type TeamSetupRequest struct {
	TeamID         int64   `json:"team_id"`
	ApplicationIDs []int64 `json:"application_ids"`
}

type SaveResultsRequest struct {
	Entries []EntryRequest `json:"entries"`
}

type SaveManualResultsRequest struct {
	Revision int64                `json:"revision"`
	Entries  []ManualEntryRequest `json:"entries"`
}

type ManualEntryRequest struct {
	EntryID      int64          `json:"entry_id"`
	ResultStatus string         `json:"result_status"`
	Place        *int32         `json:"place"`
	CustomValues map[string]any `json:"custom_values"`
}

type EntryRequest struct {
	EntryID      int64         `json:"entry_id"`
	ResultStatus string        `json:"result_status"`
	Tasks        []TaskRequest `json:"tasks"`
}

type TaskRequest struct {
	TaskNumber           int16  `json:"task_number"`
	Status               string `json:"status"`
	SolveTimeDeciseconds *int64 `json:"solve_time_deciseconds"`
	NegativeAttempts     int32  `json:"negative_attempts"`
}

type TableResponse struct {
	Competition  CompetitionResponse `json:"competition"`
	Event        EventResponse       `json:"event"`
	Mode         string              `json:"mode"`
	TaskCount    int16               `json:"task_count"`
	Status       string              `json:"status"`
	CalculatedAt *time.Time          `json:"calculated_at"`
	ApprovedAt   *time.Time          `json:"approved_at"`
	ProtocolURL  *string             `json:"protocol_url"`
	Entries      []EntryResponse     `json:"entries"`
	Revision     int64               `json:"revision"`
	ScoringMode  string              `json:"scoring_mode"`
	TemplateID   *int64              `json:"template_id"`
	Template     *TemplateSnapshot   `json:"template_snapshot"`
}

type CompetitionResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type EventResponse struct {
	ID             int64   `json:"id"`
	DisciplineID   int64   `json:"discipline_id"`
	DisciplineName string  `json:"discipline_name"`
	AgeGroupName   *string `json:"age_group_name"`
}

type EntryResponse struct {
	ID                       int64                 `json:"id"`
	Team                     *TeamResponse         `json:"team"`
	Participants             []ParticipantResponse `json:"participants"`
	ResultStatus             string                `json:"result_status"`
	Tasks                    []TaskResponse        `json:"tasks"`
	SolvedCount              int32                 `json:"solved_count"`
	TotalTimeDeciseconds     int64                 `json:"total_time_deciseconds"`
	LastSolveTimeDeciseconds int64                 `json:"last_solve_time_deciseconds"`
	NegativeAttempts         int32                 `json:"negative_attempts"`
	Place                    *int32                `json:"place"`
	CustomValues             map[string]any        `json:"custom_values"`
}

type CreateTemplateRequest struct {
	DisciplineID int64           `json:"discipline_id"`
	Name         string          `json:"name"`
	Fields       []TemplateField `json:"fields"`
}

type UpdateTemplateRequest struct {
	Name     *string          `json:"name"`
	Fields   *[]TemplateField `json:"fields"`
	IsActive *bool            `json:"is_active"`
}

type TeamResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ParticipantResponse struct {
	ApplicationID int64   `json:"application_id"`
	UserID        int64   `json:"user_id"`
	LastName      string  `json:"last_name"`
	FirstName     string  `json:"first_name"`
	MiddleName    *string `json:"middle_name"`
}

type TaskResponse struct {
	TaskNumber           int16  `json:"task_number"`
	Status               string `json:"status"`
	SolveTimeDeciseconds *int64 `json:"solve_time_deciseconds"`
	NegativeAttempts     int32  `json:"negative_attempts"`
}

type ApprovalResponse struct {
	Status      string    `json:"status"`
	ApprovedAt  time.Time `json:"approved_at"`
	ProtocolURL string    `json:"protocol_url"`
}

type HistoryItem struct {
	ResultID        int64     `json:"result_id"`
	CompetitionID   int64     `json:"competition_id"`
	CompetitionName string    `json:"competition_name"`
	EventID         int64     `json:"event_id"`
	DisciplineID    int64     `json:"discipline_id"`
	DisciplineName  string    `json:"discipline_name"`
	StartAt         time.Time `json:"start_at"`
	Place           *int32    `json:"place"`
	Score           *float64  `json:"score"`
	ResultStatus    *string   `json:"result_status"`
	ProtocolURL     *string   `json:"protocol_url"`
}
