package result

import "testing"

func solvedTask(number int16, time int64, negatives int32) TaskResult {
	return TaskResult{
		TaskNumber: number, Status: TaskSolved,
		SolveTimeDeciseconds: &time, NegativeAttempts: negatives,
	}
}

func TestCalculateAlgorithmicUsesOfficialCriteriaAndSharedPlaces(t *testing.T) {
	entries := []Entry{
		{ID: 1, ResultStatus: ResultFinished, Tasks: []TaskResult{solvedTask(1, 100, 2)}},
		{ID: 2, ResultStatus: ResultFinished, Tasks: []TaskResult{solvedTask(1, 80, 5)}},
		{ID: 3, ResultStatus: ResultFinished, Tasks: []TaskResult{solvedTask(1, 80, 5)}},
		{ID: 4, ResultStatus: ResultFinished, Tasks: nil},
		{ID: 5, ResultStatus: ResultDidNotStart},
		{ID: 6, ResultStatus: ResultDisqualified},
	}
	summaries := calculateAlgorithmic(entries)
	if len(summaries) != 6 {
		t.Fatalf("summary count = %d", len(summaries))
	}
	if summaries[0].EntryID != 2 || *summaries[0].Place != 1 ||
		summaries[1].EntryID != 3 || *summaries[1].Place != 1 {
		t.Fatalf("shared first place = %#v", summaries[:2])
	}
	if summaries[2].EntryID != 1 || *summaries[2].Place != 3 {
		t.Fatalf("third ranked entry = %#v", summaries[2])
	}
	if summaries[3].EntryID != 4 || summaries[3].SolvedCount != 0 || *summaries[3].Place != 4 {
		t.Fatalf("zero-solved entry = %#v", summaries[3])
	}
	if summaries[4].EntryID != 5 || summaries[4].Place != nil ||
		summaries[5].EntryID != 6 || summaries[5].Place != nil {
		t.Fatalf("special statuses = %#v", summaries[4:])
	}
}

func TestCalculateAlgorithmicUsesLastSolveThenNegativeDecisions(t *testing.T) {
	entries := []Entry{
		{ID: 1, ResultStatus: ResultFinished, Tasks: []TaskResult{
			solvedTask(1, 100, 0), solvedTask(2, 200, 2),
		}},
		{ID: 2, ResultStatus: ResultFinished, Tasks: []TaskResult{
			solvedTask(1, 150, 1), solvedTask(2, 150, 0),
		}},
	}
	summaries := calculateAlgorithmic(entries)
	if summaries[0].EntryID != 2 || summaries[0].LastSolveTimeDeciseconds != 150 {
		t.Fatalf("ranked summaries = %#v", summaries)
	}
}
