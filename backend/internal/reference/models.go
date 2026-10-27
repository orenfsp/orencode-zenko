package reference

type NamedItem struct {
	ID   int64
	Name string
}

type CompetitionLevel struct {
	ID   int64
	Code *string
	Name *string
}

type AgeGroup struct {
	ID     int64
	Name   string
	MinAge *int16
	MaxAge *int16
}

type NamedItemDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type CompetitionLevelDTO struct {
	ID   int64   `json:"id"`
	Code *string `json:"code"`
	Name *string `json:"name"`
}

type AgeGroupDTO struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	MinAge *int16 `json:"min_age"`
	MaxAge *int16 `json:"max_age"`
}

type SuccessResponse struct {
	Data any `json:"data"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
