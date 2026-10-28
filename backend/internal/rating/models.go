package rating

import "time"

const (
	PermissionRead   = "rating.read"
	PermissionManage = "rating.manage"
	StatusDraft      = "draft"
	StatusLocked     = "locked"
)

type Filter struct {
	Year           int32
	DisciplineID   *int64
	MunicipalityID *int64
	AgeGroupID     *int64
	Limit          int32
	Offset         int32
}

type Entry struct {
	Position         int64   `json:"position"`
	UserID           int64   `json:"user_id"`
	LastName         string  `json:"last_name"`
	FirstName        string  `json:"first_name"`
	MiddleName       *string `json:"middle_name,omitempty"`
	MunicipalityName *string `json:"municipality_name,omitempty"`
	TotalPoints      float64 `json:"total_points"`
	ResultCount      int64   `json:"result_count"`
	Wins             int64   `json:"wins"`
}

type Page struct {
	Items  []Entry `json:"items"`
	Limit  int32   `json:"limit"`
	Offset int32   `json:"offset"`
}

type Reference struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type SelectionMember struct {
	SelectionPosition int32   `json:"selection_position"`
	UserID            int64   `json:"user_id"`
	LastName          string  `json:"last_name"`
	FirstName         string  `json:"first_name"`
	MiddleName        *string `json:"middle_name,omitempty"`
}

type Selection struct {
	ID              int64             `json:"id"`
	SeasonYear      int32             `json:"season_year"`
	Discipline      Reference         `json:"discipline"`
	AgeGroup        Reference         `json:"age_group"`
	Status          string            `json:"status"`
	CaptainUserID   *int64            `json:"captain_user_id"`
	CoachUserID     *int64            `json:"coach_user_id"`
	Members         []SelectionMember `json:"members"`
	CreatedByUserID int64             `json:"created_by_user_id"`
	LockedAt        *time.Time        `json:"locked_at"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

type SelectionScope struct {
	SeasonYear   int32
	DisciplineID int64
	AgeGroupID   int64
}

type SaveSelectionRequest struct {
	SeasonYear    int32   `json:"season_year"`
	DisciplineID  int64   `json:"discipline_id"`
	AgeGroupID    int64   `json:"age_group_id"`
	MemberIDs     []int64 `json:"member_ids"`
	CaptainUserID *int64  `json:"captain_user_id"`
	CoachUserID   *int64  `json:"coach_user_id"`
}
