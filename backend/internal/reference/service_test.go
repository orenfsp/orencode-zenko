package reference

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	disciplines      []NamedItem
	competitionTypes []NamedItem
	levels           []CompetitionLevel
	ageGroups        []AgeGroup
	err              error
}

func (r *repositoryStub) ListDisciplines(context.Context) ([]NamedItem, error) {
	return r.disciplines, r.err
}
func (r *repositoryStub) ListCompetitionTypes(context.Context) ([]NamedItem, error) {
	return r.competitionTypes, r.err
}
func (r *repositoryStub) ListCompetitionLevels(context.Context) ([]CompetitionLevel, error) {
	return r.levels, r.err
}
func (r *repositoryStub) ListAgeGroups(context.Context) ([]AgeGroup, error) {
	return r.ageGroups, r.err
}

func TestServiceListsReferenceData(t *testing.T) {
	code := "regional"
	levelName := "Региональный"
	minAge, maxAge := int16(14), int16(17)
	repository := &repositoryStub{
		disciplines:      []NamedItem{{ID: 1, Name: "программирование алгоритмическое"}},
		competitionTypes: []NamedItem{{ID: 2, Name: "Чемпионат"}},
		levels:           []CompetitionLevel{{ID: 3, Code: &code, Name: &levelName}},
		ageGroups:        []AgeGroup{{ID: 4, Name: "Юниоры", MinAge: &minAge, MaxAge: &maxAge}},
	}
	service := NewService(repository)

	disciplines, err := service.ListDisciplines(context.Background())
	if err != nil || len(disciplines) != 1 || disciplines[0].Name != repository.disciplines[0].Name {
		t.Fatalf("ListDisciplines() = %#v, %v", disciplines, err)
	}
	types, err := service.ListCompetitionTypes(context.Background())
	if err != nil || len(types) != 1 || types[0].ID != 2 {
		t.Fatalf("ListCompetitionTypes() = %#v, %v", types, err)
	}
	levels, err := service.ListCompetitionLevels(context.Background())
	if err != nil || len(levels) != 1 || levels[0].Code == nil || *levels[0].Code != code {
		t.Fatalf("ListCompetitionLevels() = %#v, %v", levels, err)
	}
	ageGroups, err := service.ListAgeGroups(context.Background())
	if err != nil || len(ageGroups) != 1 || ageGroups[0].MinAge == nil || *ageGroups[0].MinAge != minAge {
		t.Fatalf("ListAgeGroups() = %#v, %v", ageGroups, err)
	}
}

func TestServiceReturnsNonNilEmptySlices(t *testing.T) {
	service := NewService(&repositoryStub{})
	checks := []struct {
		name string
		list func() (int, bool, error)
	}{
		{"disciplines", func() (int, bool, error) {
			v, e := service.ListDisciplines(context.Background())
			return len(v), v == nil, e
		}},
		{"competition types", func() (int, bool, error) {
			v, e := service.ListCompetitionTypes(context.Background())
			return len(v), v == nil, e
		}},
		{"competition levels", func() (int, bool, error) {
			v, e := service.ListCompetitionLevels(context.Background())
			return len(v), v == nil, e
		}},
		{"age groups", func() (int, bool, error) {
			v, e := service.ListAgeGroups(context.Background())
			return len(v), v == nil, e
		}},
	}
	for _, check := range checks {
		t.Run(check.name, func(t *testing.T) {
			length, isNil, err := check.list()
			if err != nil || isNil || length != 0 {
				t.Fatalf("list = length %d, nil %v, error %v", length, isNil, err)
			}
		})
	}
}

func TestServicePropagatesRepositoryErrors(t *testing.T) {
	wantErr := errors.New("database unavailable")
	service := NewService(&repositoryStub{err: wantErr})
	checks := []func() error{
		func() error { _, err := service.ListDisciplines(context.Background()); return err },
		func() error { _, err := service.ListCompetitionTypes(context.Background()); return err },
		func() error { _, err := service.ListCompetitionLevels(context.Background()); return err },
		func() error { _, err := service.ListAgeGroups(context.Background()); return err },
	}
	for _, check := range checks {
		if err := check(); !errors.Is(err, wantErr) {
			t.Fatalf("error = %v, want %v", err, wantErr)
		}
	}
}
