package main

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type upserterStub struct {
	names []string
	errAt int
}

func (s *upserterStub) UpsertDiscipline(_ context.Context, name string) error {
	if len(s.names) == s.errAt && s.errAt >= 0 {
		return errors.New("database unavailable")
	}
	s.names = append(s.names, name)
	return nil
}

func TestSeedDisciplinesUsesOfficialValues(t *testing.T) {
	stub := &upserterStub{errAt: -1}
	if err := seedDisciplines(context.Background(), stub); err != nil {
		t.Fatalf("seedDisciplines() error = %v", err)
	}
	if !reflect.DeepEqual(stub.names, officialDisciplines) || len(stub.names) != 6 {
		t.Fatalf("seeded disciplines = %#v", stub.names)
	}
	unique := make(map[string]struct{}, len(officialDisciplines))
	for _, name := range officialDisciplines {
		if name == "" {
			t.Fatal("official discipline name must not be empty")
		}
		if _, exists := unique[name]; exists {
			t.Fatalf("duplicate official discipline %q", name)
		}
		unique[name] = struct{}{}
	}
}

func TestSeedDisciplinesStopsOnError(t *testing.T) {
	stub := &upserterStub{errAt: 2}
	if err := seedDisciplines(context.Background(), stub); err == nil {
		t.Fatal("seedDisciplines() expected error")
	}
	if len(stub.names) != 2 {
		t.Fatalf("successful upserts = %d, want 2", len(stub.names))
	}
}
