package result

import "context"

type Repository interface {
	GetEventContext(ctx context.Context, competitionID, eventID int64) (EventContext, error)
	ListApprovedApplications(ctx context.Context, eventID int64) ([]Application, error)
	ListEligibleTeamMemberships(ctx context.Context, eventID int64) ([]TeamMembership, error)
	Configure(ctx context.Context, eventID int64, mode string, taskCount int16, teams []TeamSetup, confirmReset bool) error
	GetTable(ctx context.Context, eventID int64) (Table, error)
	SaveDraftEntries(ctx context.Context, eventID int64, entries []EntryUpdate) error
	SaveCalculated(ctx context.Context, eventID, version int64, summaries []EntrySummary) error
	Approve(ctx context.Context, eventID, actorUserID, version int64, summaries []EntrySummary, protocolKey, protocolURL string) (Approval, error)
	ListProtocolJudges(ctx context.Context, competitionID int64) ([]Judge, error)
	ListHistory(ctx context.Context, userID int64, disciplineID *int64) ([]HistoryItem, error)
	ReplaceProtocol(ctx context.Context, eventID int64, key, url *string) (*string, error)
}
