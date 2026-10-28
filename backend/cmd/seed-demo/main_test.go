package main

import (
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/result"
)

func TestDemoDisciplineMatchesResultRules(t *testing.T) {
	if demoDisciplineName != result.AlgorithmicDiscipline {
		t.Fatalf("demo discipline = %q, want %q", demoDisciplineName, result.AlgorithmicDiscipline)
	}
}
