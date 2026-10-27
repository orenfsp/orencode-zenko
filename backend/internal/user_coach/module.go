package usercoach

import "time"

type Status string

const (
    StatusPending   Status = "pending"
    StatusConfirmed Status = "confirmed"
    StatusEnded     Status = "ended"
)

type UserCoach struct {
    ID           int64
    UserID       int64 // ID участника
    CoachUserID  int64 // ID тренера
    DisciplineID int64
    Status       Status
    RequestedAt  time.Time
    ConfirmedAt  *time.Time
    EndedAt      *time.Time
}