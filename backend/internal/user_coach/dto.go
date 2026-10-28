package usercoach

import "time"

type ParticipantDiscipline struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ParticipantRelation struct {
	DisciplineID int64  `json:"discipline_id"`
	Status       string `json:"status"`
}

type Participant struct {
	ID             int64                   `json:"id"`
	FirstName      string                  `json:"first_name"`
	LastName       string                  `json:"last_name"`
	MiddleName     *string                 `json:"middle_name"`
	Organization   *string                 `json:"organization"`
	Municipality   *string                 `json:"municipality"`
	Disciplines    []ParticipantDiscipline `json:"disciplines"`
	CoachRelations []ParticipantRelation   `json:"coach_relations"`
}

type ParticipantSearchResponse struct {
	Items  []Participant `json:"items"`
	Total  int64         `json:"total"`
	Limit  int32         `json:"limit"`
	Offset int32         `json:"offset"`
}

type CreateRequest struct {
	ParticipantUserID int64 `json:"participant_user_id"`
	DisciplineID      int64 `json:"discipline_id"`
}

type Response struct {
	ID                int64      `json:"id"`
	ParticipantUserID int64      `json:"participant_user_id"`
	CoachUserID       int64      `json:"coach_user_id"`
	DisciplineID      int64      `json:"discipline_id"`
	Status            Status     `json:"status"`
	RequestedAt       time.Time  `json:"requested_at"`
	ConfirmedAt       *time.Time `json:"confirmed_at,omitempty"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
}
