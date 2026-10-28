package dashboard

import (
	"errors"
	"time"
)

var ErrInvalidYear = errors.New("invalid dashboard year")

const (
	PermissionRead = "dashboard.read"
	defaultLimit   = int32(10)
	expiringDays   = int32(60)
)

type Metrics struct {
	Athletes     int64 `json:"athletes"`
	Competitions int64 `json:"competitions"`
	Participants int64 `json:"participants"`
	ActiveRanks  int64 `json:"active_ranks"`
}

type UpcomingCompetition struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	Status         string    `json:"status"`
	DisciplineName string    `json:"discipline_name"`
}

type PendingApplication struct {
	ID              int64     `json:"id"`
	SubmittedAt     time.Time `json:"submitted_at"`
	AthleteUserID   int64     `json:"athlete_user_id"`
	AthleteName     string    `json:"athlete_name"`
	CompetitionID   *int64    `json:"competition_id,omitempty"`
	CompetitionName *string   `json:"competition_name,omitempty"`
	RankName        *string   `json:"rank_name,omitempty"`
	DisciplineName  *string   `json:"discipline_name,omitempty"`
	CategoryName    *string   `json:"category_name,omitempty"`
	ApplicationType *string   `json:"application_type,omitempty"`
}

type ExpiringCredential struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	UserName       string    `json:"user_name"`
	CredentialName string    `json:"credential_name"`
	DisciplineName *string   `json:"discipline_name,omitempty"`
	ExpiresAt      time.Time `json:"expires_at"`
}

type Deadline struct {
	Type          string    `json:"type"`
	Title         string    `json:"title"`
	DueAt         time.Time `json:"due_at"`
	CompetitionID *int64    `json:"competition_id,omitempty"`
}

type Data struct {
	PeriodYear                     int                   `json:"period_year"`
	Metrics                        Metrics               `json:"metrics"`
	UpcomingCompetitions           []UpcomingCompetition `json:"upcoming_competitions"`
	PendingCompetitionApplications []PendingApplication  `json:"pending_competition_applications"`
	PendingRankApplications        []PendingApplication  `json:"pending_rank_applications"`
	PendingJudgeApplications       []PendingApplication  `json:"pending_judge_category_applications"`
	ExpiringRanks                  []ExpiringCredential  `json:"expiring_ranks"`
	ExpiringJudgeCategories        []ExpiringCredential  `json:"expiring_judge_categories"`
	Deadlines                      []Deadline            `json:"deadlines"`
}
