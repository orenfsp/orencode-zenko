package rankuser

import "time"

type CreateRequest struct {
	AthleteUserID int64 `json:"athlete_user_id,omitempty"`
	RankID        int64 `json:"rank_id"`
	DisciplineID  int64 `json:"discipline_id"`
}

type ReviewRequest struct {
	Status          string `json:"status"`
	RejectionReason string `json:"rejection_reason,omitempty"`
}

type Response struct {
	ID                int64      `json:"id"`
	AthleteUserID     int64      `json:"athlete_user_id"`
	RankID            int64      `json:"rank_id"`
	DisciplineID      int64      `json:"discipline_id"`
	SubmittedByUserID int64      `json:"submitted_by_user_id"`
	Status            string     `json:"status"`
	ReviewedByUserID  *int64     `json:"reviewed_by_user_id,omitempty"`
	ReviewedAt        *time.Time `json:"reviewed_at,omitempty"`
	RejectionReason   *string    `json:"rejection_reason,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}
