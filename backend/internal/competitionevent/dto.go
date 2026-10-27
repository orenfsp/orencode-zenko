package competitionevent

import "github.com/orenfsp/orencode-zenko.git/internal/competition"

type CreateRequest struct {
	AgeGroupID *int64 `json:"age_group_id"`
}

type UpdateRequest struct {
	AgeGroupID competition.PatchField[int64] `json:"age_group_id"`
}

type ReferenceDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ResponseDTO struct {
	ID            int64         `json:"id"`
	CompetitionID int64         `json:"competition_id"`
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
