package rating

import (
	"bytes"
	"testing"
	"time"
)

func TestPDFOrderGeneratorProducesPDF(t *testing.T) {
	content, err := NewPDFOrderGenerator().Generate(Selection{
		SeasonYear: 2026,
		Discipline: Reference{Name: "Программирование алгоритмическое"},
		AgeGroup:   Reference{Name: "Юниоры"},
		Members:    []SelectionMember{{SelectionPosition: 1, UserID: 10, LastName: "Иванов", FirstName: "Иван"}},
	}, nil, time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(content, []byte("%PDF-")) || !bytes.Contains(content[len(content)-128:], []byte("%%EOF")) {
		t.Fatal("generator did not produce a valid PDF envelope")
	}
}
