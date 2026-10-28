package result

import "testing"

func TestValidateManualUpdates(t *testing.T) {
	place1, place2 := int32(1), int32(2)
	table := Table{Sheet: Sheet{TemplateSnapshot: &TemplateSnapshot{Fields: []TemplateField{{Key: "score", Label: "Оценка", Type: "number", Required: true}}}}, Entries: []Entry{{ID: 1}, {ID: 2}}}
	request := SaveManualResultsRequest{Entries: []ManualEntryRequest{
		{EntryID: 1, ResultStatus: ResultFinished, Place: &place1, CustomValues: map[string]any{"score": 95.5}},
		{EntryID: 2, ResultStatus: ResultFinished, Place: &place2, CustomValues: map[string]any{"score": 90.0}},
	}}
	if _, err := validateManualUpdates(table, request, true); err != nil {
		t.Fatalf("valid manual results: %v", err)
	}
	request.Entries[1].Place = &place1
	if _, err := validateManualUpdates(table, request, true); err != ErrInvalidRequest {
		t.Fatalf("duplicate place error = %v", err)
	}
	request.Entries[1].Place = &place2
	request.Entries[1].CustomValues = map[string]any{}
	if _, err := validateManualUpdates(table, request, true); err != ErrInvalidRequest {
		t.Fatalf("required field error = %v", err)
	}
}
