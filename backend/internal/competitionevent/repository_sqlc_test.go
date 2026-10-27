package competitionevent

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapWriteError(t *testing.T) {
	tests := []struct {
		constraint string
		want       error
	}{
		{constraint: "competition_event_age_group_id_fkey", want: ErrAgeGroupMissing},
		{constraint: "competition_application_event_id_fkey", want: ErrHasApplications},
		{constraint: "competition_event_competition_id_fkey", want: ErrCompetitionMissing},
	}
	for _, tt := range tests {
		err := &pgconn.PgError{Code: "23503", ConstraintName: tt.constraint}
		if got := mapWriteError(err); !errors.Is(got, tt.want) {
			t.Fatalf("mapWriteError(%q) = %v, want %v", tt.constraint, got, tt.want)
		}
	}
}

func TestMapWriteErrorPreservesUnknownDatabaseError(t *testing.T) {
	want := errors.New("database unavailable")
	if got := mapWriteError(want); !errors.Is(got, want) {
		t.Fatalf("mapWriteError() = %v, want original error", got)
	}
}

func TestSameNullableID(t *testing.T) {
	one, anotherOne, two := int64(1), int64(1), int64(2)
	if !sameNullableID(nil, nil) || !sameNullableID(&one, &anotherOne) || sameNullableID(&one, &two) || sameNullableID(&one, nil) {
		t.Fatal("sameNullableID returned an unexpected result")
	}
}
