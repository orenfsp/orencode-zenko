package result

import "sort"

// calculateAlgorithmic follows section 5.20.1 of the official rules:
// solved tasks (descending), total solve time, last solve time and negative
// decisions (all ascending). Exact ties share a place (competition ranking).
func calculateAlgorithmic(entries []Entry) []EntrySummary {
	type ranked struct {
		EntrySummary
		status string
	}
	items := make([]ranked, 0, len(entries))
	for _, entry := range entries {
		summary := EntrySummary{EntryID: entry.ID}
		if entry.ResultStatus == ResultFinished {
			for _, task := range entry.Tasks {
				summary.NegativeAttempts += task.NegativeAttempts
				if task.Status != TaskSolved || task.SolveTimeDeciseconds == nil {
					continue
				}
				summary.SolvedCount++
				summary.TotalTimeDeciseconds += *task.SolveTimeDeciseconds
				if *task.SolveTimeDeciseconds > summary.LastSolveTimeDeciseconds {
					summary.LastSolveTimeDeciseconds = *task.SolveTimeDeciseconds
				}
			}
		}
		items = append(items, ranked{EntrySummary: summary, status: entry.ResultStatus})
	}

	sort.SliceStable(items, func(i, j int) bool {
		left, right := items[i], items[j]
		if statusOrder(left.status) != statusOrder(right.status) {
			return statusOrder(left.status) < statusOrder(right.status)
		}
		if left.status != ResultFinished {
			return left.EntryID < right.EntryID
		}
		if left.SolvedCount != right.SolvedCount {
			return left.SolvedCount > right.SolvedCount
		}
		if left.TotalTimeDeciseconds != right.TotalTimeDeciseconds {
			return left.TotalTimeDeciseconds < right.TotalTimeDeciseconds
		}
		if left.LastSolveTimeDeciseconds != right.LastSolveTimeDeciseconds {
			return left.LastSolveTimeDeciseconds < right.LastSolveTimeDeciseconds
		}
		if left.NegativeAttempts != right.NegativeAttempts {
			return left.NegativeAttempts < right.NegativeAttempts
		}
		return left.EntryID < right.EntryID
	})

	for index := range items {
		if items[index].status != ResultFinished {
			continue
		}
		place := int32(index + 1)
		if index > 0 && equalRank(items[index-1], items[index]) {
			place = *items[index-1].Place
		}
		items[index].Place = &place
	}
	result := make([]EntrySummary, 0, len(items))
	for _, item := range items {
		result = append(result, item.EntrySummary)
	}
	return result
}

func statusOrder(status string) int {
	switch status {
	case ResultFinished:
		return 0
	case ResultDidNotStart:
		return 1
	default:
		return 2
	}
}

func equalRank(left, right struct {
	EntrySummary
	status string
}) bool {
	if left.status != right.status {
		return false
	}
	if left.status != ResultFinished {
		return true
	}
	return left.SolvedCount == right.SolvedCount &&
		left.TotalTimeDeciseconds == right.TotalTimeDeciseconds &&
		left.LastSolveTimeDeciseconds == right.LastSolveTimeDeciseconds &&
		left.NegativeAttempts == right.NegativeAttempts
}
