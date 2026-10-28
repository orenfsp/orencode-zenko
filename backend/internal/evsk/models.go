package evsk

import "time"

type Standard struct {
	ID           int64      `json:"id"`
	RankID       int64      `json:"rank_id"`
	DisciplineID int64      `json:"discipline_id"`
	MinimumPlace *int32     `json:"minimum_place,omitempty"`
	MinimumScore *string    `json:"minimum_score,omitempty"`
	ValidFrom    *time.Time `json:"valid_from,omitempty"`
	ValidTo      *time.Time `json:"valid_to,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
type UpsertRequest struct {
	RankID       int64   `json:"rank_id"`
	DisciplineID int64   `json:"discipline_id"`
	MinimumPlace *int32  `json:"minimum_place,omitempty"`
	MinimumScore *string `json:"minimum_score,omitempty"`
	ValidFrom    *string `json:"valid_from,omitempty"`
	ValidTo      *string `json:"valid_to,omitempty"`
}
