package calendarplan

import (
	"bytes"
	"encoding/json"
	"time"
)

type CreateRequest struct {
	Name string `json:"name"`
	Year int16  `json:"year"`
}

// PatchField distinguishes an omitted JSON field from an explicit null.
type PatchField[T any] struct {
	Present bool
	Null    bool
	Value   T
}

func (f *PatchField[T]) UnmarshalJSON(data []byte) error {
	f.Present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		f.Null = true
		var zero T
		f.Value = zero
		return nil
	}
	f.Null = false
	return json.Unmarshal(data, &f.Value)
}

type UpdateRequest struct {
	Name PatchField[string] `json:"name"`
	Year PatchField[int16]  `json:"year"`
}

func (r UpdateRequest) Empty() bool {
	return !r.Name.Present && !r.Year.Present
}

type ResponseDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Year int16  `json:"year"`
}

type ReferenceDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CompetitionDTO struct {
	ID                 int64        `json:"id"`
	Name               string       `json:"name"`
	StartAt            time.Time    `json:"start_at"`
	EndAt              time.Time    `json:"end_at"`
	Status             string       `json:"status"`
	Discipline         ReferenceDTO `json:"discipline"`
	CalendarPlanNumber *string      `json:"calendar_plan_number"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}
