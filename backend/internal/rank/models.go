package rank

import (
	"time"
)
type Create struct {
	Name string
	TTL  int
}

type Response struct {
	ID      int32
	Name  	string
	TTL     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

