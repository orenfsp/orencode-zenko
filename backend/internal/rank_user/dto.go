package rankuser

import "time"

type CreateRequest struct {
	AthleteUserID   int64  `json:"athlete_user_id,omitempty"`
	RankID          int64  `json:"rank_id"`
	DisciplineID    int64  `json:"discipline_id"`
	ApplicationType string `json:"application_type"`
	CompetitionID   int64  `json:"competition_id"`
	ResultID        int64  `json:"result_id"`
}

type ReviewRequest struct {
	Status          string  `json:"status"`
	RejectionReason string  `json:"rejection_reason,omitempty"`
	AssignedAt      *string `json:"assigned_at"`
	ExpiresAt       *string `json:"expires_at"`
	OrderNumber     *string `json:"order_number"`
}

type Response struct {
	ID                int64      `json:"id"`
	AthleteUserID     int64      `json:"athlete_user_id"`
	RankID            int64      `json:"rank_id"`
	DisciplineID      int64      `json:"discipline_id"`
	ApplicationType   string     `json:"application_type"`
	CompetitionID     *int64     `json:"competition_id"`
	ResultID          *int64     `json:"result_id"`
	SubmittedByUserID int64      `json:"submitted_by_user_id"`
	Status            string     `json:"status"`
	ReviewedByUserID  *int64     `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason   *string    `json:"rejection_reason,omitempty"`
	AssignedAt        *time.Time `json:"assigned_at"`
	ExpiresAt         *time.Time `json:"expires_at"`
	OrderNumber       *string    `json:"order_number"`
	CreatedAt         time.Time  `json:"created_at"`
}

type HistoryResponse struct {
	ID              int64      `json:"id"`
	RankID          int64      `json:"rank_id"`
	RankName        string     `json:"rank_name"`
	DisciplineID    int64      `json:"discipline_id"`
	DisciplineName  string     `json:"discipline_name"`
	ResultID        *int64     `json:"result_id"`
	AwardedAt       *time.Time `json:"awarded_at"`
	ExpiresAt       *time.Time `json:"expires_at"`
	OrderNumber     *string    `json:"order_number"`
	CompetitionID   *int64     `json:"competition_id"`
	CompetitionName *string    `json:"competition_name"`
}
