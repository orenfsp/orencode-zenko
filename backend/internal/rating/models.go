package rating

const PermissionRead = "rating.read"

type Filter struct {
	Year           int32
	DisciplineID   *int64
	MunicipalityID *int64
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
