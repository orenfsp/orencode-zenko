package evsk

import (
	"errors"
	"testing"
)

func TestOptionalNumeric(t *testing.T) {
	tests := []struct {
		name  string
		value *string
		valid bool
	}{
		{name: "absent", value: nil, valid: false},
		{name: "integer", value: stringPointer("1"), valid: true},
		{name: "decimal", value: stringPointer("80.5"), valid: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			numeric, err := optionalNumeric(test.value)
			if err != nil {
				t.Fatalf("optionalNumeric() error = %v", err)
			}
			if numeric.Valid != test.valid {
				t.Fatalf("numeric.Valid = %v, want %v", numeric.Valid, test.valid)
			}
		})
	}
}

func TestOptionalNumericRejectsInvalidValue(t *testing.T) {
	_, err := optionalNumeric(stringPointer("not-a-number"))
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("optionalNumeric() error = %v, want ErrInvalidInput", err)
	}
}

func stringPointer(value string) *string { return &value }
