package competition

import (
	"bytes"
	"encoding/json"
	"time"
)

type CreateRequest struct {
	Name                 string    `json:"name"`
	Description          *string   `json:"description"`
	DisciplineID         int64     `json:"discipline_id"`
	CompetitionTypeID    int64     `json:"competition_type_id"`
	CompetitionLevelID   int64     `json:"competition_level_id"`
	Format               string    `json:"format"`
	StartAt              time.Time `json:"start_at"`
	EndAt                time.Time `json:"end_at"`
	RegistrationStartsAt time.Time `json:"registration_starts_at"`
	RegistrationEndsAt   time.Time `json:"registration_ends_at"`
	VenueAddress         *string   `json:"venue_address"`
	OnlineURL            *string   `json:"online_url"`
	CalendarPlanID       *int64    `json:"calendar_plan_id"`
	CalendarPlanNumber   *string   `json:"calendar_plan_number"`
	RegulationURL        *string   `json:"regulation_url"`
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
	Name                 PatchField[string]    `json:"name"`
	Description          PatchField[string]    `json:"description"`
	DisciplineID         PatchField[int64]     `json:"discipline_id"`
	CompetitionTypeID    PatchField[int64]     `json:"competition_type_id"`
	CompetitionLevelID   PatchField[int64]     `json:"competition_level_id"`
	Format               PatchField[string]    `json:"format"`
	StartAt              PatchField[time.Time] `json:"start_at"`
	EndAt                PatchField[time.Time] `json:"end_at"`
	RegistrationStartsAt PatchField[time.Time] `json:"registration_starts_at"`
	RegistrationEndsAt   PatchField[time.Time] `json:"registration_ends_at"`
	VenueAddress         PatchField[string]    `json:"venue_address"`
	OnlineURL            PatchField[string]    `json:"online_url"`
	CalendarPlanID       PatchField[int64]     `json:"calendar_plan_id"`
	CalendarPlanNumber   PatchField[string]    `json:"calendar_plan_number"`
	RegulationURL        PatchField[string]    `json:"regulation_url"`
}

func (r UpdateRequest) Empty() bool {
	return !r.Name.Present && !r.Description.Present && !r.DisciplineID.Present &&
		!r.CompetitionTypeID.Present && !r.CompetitionLevelID.Present && !r.Format.Present &&
		!r.StartAt.Present && !r.EndAt.Present && !r.RegistrationStartsAt.Present &&
		!r.RegistrationEndsAt.Present && !r.VenueAddress.Present && !r.OnlineURL.Present &&
		!r.CalendarPlanID.Present && !r.CalendarPlanNumber.Present && !r.RegulationURL.Present
}

type ReferenceDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CompetitionLevelDTO struct {
	ID   int64   `json:"id"`
	Code *string `json:"code"`
	Name *string `json:"name"`
}

type CalendarPlanDTO struct {
	ID     int64   `json:"id"`
	Name   *string `json:"name"`
	Year   *int16  `json:"year"`
	Number *string `json:"number"`
}

type ResponseDTO struct {
	ID                   int64               `json:"id"`
	Name                 string              `json:"name"`
	Description          *string             `json:"description"`
	Discipline           ReferenceDTO        `json:"discipline"`
	CompetitionType      ReferenceDTO        `json:"competition_type"`
	CompetitionLevel     CompetitionLevelDTO `json:"competition_level"`
	Format               string              `json:"format"`
	StartAt              time.Time           `json:"start_at"`
	EndAt                time.Time           `json:"end_at"`
	RegistrationStartsAt time.Time           `json:"registration_starts_at"`
	RegistrationEndsAt   time.Time           `json:"registration_ends_at"`
	VenueAddress         *string             `json:"venue_address"`
	OnlineURL            *string             `json:"online_url"`
	Status               string              `json:"status"`
	CalendarPlan         *CalendarPlanDTO    `json:"calendar_plan"`
	RegulationURL        *string             `json:"regulation_url"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}
