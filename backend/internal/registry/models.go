package registry

import (
	"errors"
	"time"
)

const (
	PermissionRead = "registry.read"
	DefaultLimit   = int32(50)
	MaxLimit       = int32(100)
)

var ErrInvalidFilter = errors.New("invalid registry filter")

type Reference struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Role struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Rank struct {
	ID             int64   `json:"id"`
	RankID         int64   `json:"rank_id"`
	RankName       string  `json:"rank_name"`
	DisciplineID   int64   `json:"discipline_id"`
	DisciplineName string  `json:"discipline_name"`
	AwardedAt      *string `json:"awarded_at"`
	ExpiresAt      *string `json:"expires_at"`
	OrderNumber    *string `json:"order_number"`
}

type JudgeCategory struct {
	ID           int64   `json:"id"`
	CategoryID   int64   `json:"category_id"`
	CategoryName string  `json:"category_name"`
	AssignedAt   string  `json:"assigned_at"`
	ExpiresAt    *string `json:"expires_at"`
	OrderNumber  *string `json:"order_number"`
}

type CoachRelation struct {
	ID              int64      `json:"id"`
	RelationType    string     `json:"relation_type"`
	RelatedUserID   int64      `json:"related_user_id"`
	RelatedUserName string     `json:"related_user_name"`
	Discipline      *Reference `json:"discipline,omitempty"`
	Status          string     `json:"status"`
	RequestedAt     time.Time  `json:"requested_at"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
	EndedAt         *time.Time `json:"ended_at,omitempty"`
}

type JudgeAssignment struct {
	ID                int64     `json:"id"`
	CompetitionID     int64     `json:"competition_id"`
	CompetitionName   string    `json:"competition_name"`
	CompetitionStatus string    `json:"competition_status"`
	PositionID        int64     `json:"position_id"`
	PositionName      string    `json:"position_name"`
	StartAt           time.Time `json:"start_at"`
	EndAt             time.Time `json:"end_at"`
	AppointedAt       time.Time `json:"appointed_at"`
}

type User struct {
	ID               int64             `json:"id"`
	LastName         string            `json:"last_name"`
	FirstName        string            `json:"first_name"`
	MiddleName       *string           `json:"middle_name"`
	Birth            time.Time         `json:"birth"`
	CreatedAt        time.Time         `json:"created_at"`
	Organization     *Reference        `json:"organization"`
	Municipality     *Reference        `json:"municipality"`
	Roles            []Role            `json:"roles"`
	Disciplines      []Reference       `json:"disciplines"`
	Ranks            []Rank            `json:"ranks"`
	JudgeCategory    *JudgeCategory    `json:"judge_category"`
	CoachRelations   []CoachRelation   `json:"coach_relations"`
	JudgeAssignments []JudgeAssignment `json:"judge_assignments"`
}

type Organization struct {
	ID               int64      `json:"id"`
	Name             string     `json:"name"`
	OrganizationType string     `json:"organization_type"`
	Municipality     *Reference `json:"municipality"`
}

type UserFilter struct {
	Search         *string
	DisciplineID   *int64
	OrganizationID *int64
	RankID         *int64
	AgeGroupID     *int64
	MunicipalityID *int64
	RoleCode       *string
	Limit          int32
	Offset         int32
}

type OrganizationFilter struct {
	Search         *string
	MunicipalityID *int64
	Limit          int32
	Offset         int32
}

type Page[T any] struct {
	Items  []T   `json:"items"`
	Total  int64 `json:"total"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}
