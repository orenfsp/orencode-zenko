package result

import (
	"bytes"
	"testing"
	"time"
)

func TestPDFProtocolGeneratorProducesCyrillicPDF(t *testing.T) {
	table := draftTable()
	table.Context = algorithmicContext()
	table.Sheet.TaskCount = 12
	place := int32(1)
	table.Entries[0].Place = &place
	table.Judges = []Judge{
		{PositionName: "главный судья", LastName: "Петров", FirstName: "Пётр"},
		{PositionName: "главный секретарь", LastName: "Сидорова", FirstName: "Анна"},
	}
	content, err := NewPDFProtocolGenerator().Generate(ProtocolData{Table: table, Date: time.Date(2026, 10, 27, 0, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(content) < 1000 || !bytes.HasPrefix(content, []byte("%PDF-")) {
		t.Fatalf("invalid PDF, size=%d prefix=%q", len(content), content[:min(len(content), 5)])
	}
}
