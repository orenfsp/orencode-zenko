package team

import "time"

const PermissionManage = "teams.manage"

type Member struct {
	UserID     int64   `json:"user_id"`
	LastName   string  `json:"last_name"`
	FirstName  string  `json:"first_name"`
	MiddleName *string `json:"middle_name,omitempty"`
}
type Squad struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Members   []Member  `json:"members"`
	Coaches   []Member  `json:"coaches"`
	CreatedAt time.Time `json:"created_at"`
}
type CreateRequest struct {
	Name      string  `json:"name"`
	MemberIDs []int64 `json:"member_ids"`
	CoachIDs  []int64 `json:"coach_ids"`
}
