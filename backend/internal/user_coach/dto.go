package usercoach

import "time"

type CreateRequest struct {
    ParticipantUserID int64 `json:"participant_user_id"`
    DisciplineID      int64 `json:"discipline_id"`
}

type Response struct {
    ID                  int64     `json:"id"`
    ParticipantUserID   int64     `json:"participant_user_id"`
    CoachUserID         int64     `json:"coach_user_id"`
    DisciplineID        int64     `json:"discipline_id"`
    Status              Status    `json:"status"`
    RequestedAt         time.Time `json:"requested_at"`
    ConfirmedAt         *time.Time `json:"confirmed_at,omitempty"`
    EndedAt             *time.Time `json:"ended_at,omitempty"`
}