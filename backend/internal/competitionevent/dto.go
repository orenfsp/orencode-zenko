package competitionevent

import "github.com/orenfsp/orencode-zenko.git/internal/competition"

type CreateRequest struct {
	Name       string `json:"name"`
	AgeGroupID *int64 `json:"age_group_id"`
}

type UpdateRequest struct {
	Name       competition.PatchField[string] `json:"name"`
	AgeGroupID competition.PatchField[int64]  `json:"age_group_id"`
}

type ReferenceDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ResponseDTO struct {
	ID            int64         `json:"id"`
	CompetitionID int64         `json:"competition_id"`
	Name          string        `json:"name"`
	Discipline    ReferenceDTO  `json:"discipline"`
	AgeGroup      *ReferenceDTO `json:"age_group"`
	ProtocolURL   *string       `json:"protocol_url"`
}

type errorResponse struct {
	Message string `json:"message"`
}

type successResponse struct {
	Data any `json:"data"`
}
