package result

import "testing"

func TestValidateTemplate(t *testing.T) {
	valid := TemplateField{Key: "quality", Label: "Качество", Type: "number", Required: true}
	for _, tc := range []struct {
		name    string
		fields  []TemplateField
		wantErr bool
	}{
		{"Шаблон", []TemplateField{valid}, false},
		{" ", []TemplateField{valid}, true},
		{"Шаблон", []TemplateField{valid, valid}, true},
		{"Шаблон", []TemplateField{{Key: "bad-key", Label: "Поле", Type: "text"}}, true},
		{"Шаблон", []TemplateField{{Key: "field", Label: "Поле", Type: "formula"}}, true},
		{"Шаблон", nil, true},
	} {
		if err := validateTemplate(tc.name, tc.fields); (err != nil) != tc.wantErr {
			t.Fatalf("validateTemplate(%q, %v) = %v", tc.name, tc.fields, err)
		}
	}
}
