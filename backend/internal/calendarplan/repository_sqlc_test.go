package calendarplan

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestMapReadError(t *testing.T) {
	if !errors.Is(mapReadError(pgx.ErrNoRows), ErrNotFound) {
		t.Fatal("pgx.ErrNoRows must map to ErrNotFound")
	}
	original := errors.New("database unavailable")
	if !errors.Is(mapReadError(original), original) {
		t.Fatal("unexpected repository errors must be preserved")
	}
}
