package competition

import "time"

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
