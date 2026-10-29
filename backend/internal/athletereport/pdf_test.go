package athletereport

import (
	"bytes"
	"testing"
	"time"

	"github.com/orenfsp/orencode-zenko.git/internal/auth"
)

func TestPDFGeneratorProducesPDF(t *testing.T) {
	content, err := NewPDFGenerator().Generate(Data{
		Profile:   auth.Profile{FirstName: "Иван", LastName: "Спортсменов", BirthDate: time.Date(2005, 5, 15, 0, 0, 0, 0, time.UTC)},
		CreatedAt: time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(content, []byte("%PDF-")) || !bytes.Contains(content[len(content)-128:], []byte("%%EOF")) {
		t.Fatal("generator did not produce a valid PDF envelope")
	}
}
