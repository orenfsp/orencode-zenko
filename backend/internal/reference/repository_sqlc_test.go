package reference

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestNullableSQLTypesBecomePointers(t *testing.T) {
	if textPointer(pgtype.Text{}) != nil || int16Pointer(pgtype.Int2{}) != nil {
		t.Fatal("invalid pgtype values must become nil pointers")
	}
	text := textPointer(pgtype.Text{String: "regional", Valid: true})
	age := int16Pointer(pgtype.Int2{Int16: 14, Valid: true})
	if text == nil || *text != "regional" || age == nil || *age != 14 {
		t.Fatalf("converted values = %v, %v", text, age)
	}
}
