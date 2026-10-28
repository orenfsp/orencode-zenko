package dashboard

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	data  Data
	err   error
	limit int32
	days  int32
	year  int
}

func (r *repositoryStub) Load(_ context.Context, limit, days int32, year int) (Data, error) {
	r.limit, r.days, r.year = limit, days, year
	return r.data, r.err
}

func TestServiceGetNormalizesEmptyLists(t *testing.T) {
	repository := &repositoryStub{data: Data{Metrics: Metrics{Athletes: 3}}}
	data, err := NewService(repository).Get(context.Background(), 2026)
	if err != nil {
		t.Fatal(err)
	}
	if data.Metrics.Athletes != 3 || data.UpcomingCompetitions == nil || data.Deadlines == nil {
		t.Fatalf("unexpected data: %#v", data)
	}
	if repository.limit != 10 || repository.days != 60 {
		t.Fatalf("limit=%d days=%d", repository.limit, repository.days)
	}
	if repository.year != 2026 || data.PeriodYear != 2026 {
		t.Fatalf("year=%d, data year=%d", repository.year, data.PeriodYear)
	}
}

func TestServiceGetReturnsRepositoryError(t *testing.T) {
	want := errors.New("database unavailable")
	_, err := NewService(&repositoryStub{err: want}).Get(context.Background(), 2026)
	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func TestServiceRejectsInvalidYear(t *testing.T) {
	_, err := NewService(&repositoryStub{}).Get(context.Background(), 1999)
	if !errors.Is(err, ErrInvalidYear) {
		t.Fatalf("error = %v, want ErrInvalidYear", err)
	}
}
