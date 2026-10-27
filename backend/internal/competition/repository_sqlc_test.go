package competition

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	competitionerrs "github.com/orenfsp/orencode-zenko.git/internal/competition/errs"
)

func TestMapWriteError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{
			name: "duplicate competition name",
			err:  &pgconn.PgError{Code: "23505", ConstraintName: "competition_name_key"},
			want: competitionerrs.ErrNameConflict,
		},
		{
			name: "invalid reference",
			err:  &pgconn.PgError{Code: "23503", ConstraintName: "competition_discipline_id_fkey"},
			want: competitionerrs.ErrInvalidReference,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mapWriteError(tt.err); !errors.Is(got, tt.want) {
				t.Fatalf("mapWriteError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMapWriteErrorPreservesUnknownError(t *testing.T) {
	want := errors.New("database unavailable")
	if got := mapWriteError(want); !errors.Is(got, want) {
		t.Fatalf("mapWriteError() = %v, want original error", got)
	}
}
