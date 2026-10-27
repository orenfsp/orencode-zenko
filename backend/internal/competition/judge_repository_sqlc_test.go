package competition

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestMapJudgeWriteError(t *testing.T) {
	foreignKeyError := &pgconn.PgError{Code: "23503"}
	if !errors.Is(mapJudgeWriteError(foreignKeyError), ErrInvalidJudgeReference) {
		t.Fatal("foreign key violation must map to ErrInvalidJudgeReference")
	}
	original := errors.New("database unavailable")
	if !errors.Is(mapJudgeWriteError(original), original) {
		t.Fatal("unknown repository error must be preserved")
	}
}
