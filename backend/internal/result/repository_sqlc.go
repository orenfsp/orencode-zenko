package result

import (
	"context"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/orenfsp/orencode-zenko.git/db/sqlc"
)

type SQLCRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewRepository(queries *db.Queries, pool *pgxpool.Pool) *SQLCRepository {
	return &SQLCRepository{queries: queries, pool: pool}
}

func (r *SQLCRepository) ReplaceProtocol(ctx context.Context, eventID int64, key, url *string) (*string, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	sheet, err := queries.LockEventResultSheet(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProtocolNotApproved
	}
	if err != nil {
		return nil, err
	}
	if sheet.Status != SheetApproved {
		return nil, ErrProtocolNotApproved
	}
	oldKey := textPointer(sheet.ProtocolStorageKey)
	if key == nil && oldKey == nil {
		return nil, ErrProtocolNotFound
	}
	newKey := pgtype.Text{}
	newURL := pgtype.Text{}
	if key != nil {
		newKey = pgtype.Text{String: *key, Valid: true}
	}
	if url != nil {
		newURL = pgtype.Text{String: *url, Valid: true}
	}
	if err := queries.SetResultProtocolStorageKey(ctx, db.SetResultProtocolStorageKeyParams{PSheetID: sheet.ID, PProtocolStorageKey: newKey}); err != nil {
		return nil, err
	}
	if err := queries.SetCompetitionEventProtocolURL(ctx, db.SetCompetitionEventProtocolURLParams{PEventID: eventID, PProtocolUrl: newURL}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return oldKey, nil
}

func (r *SQLCRepository) GetEventContext(
	ctx context.Context, competitionID, eventID int64,
) (EventContext, error) {
	row, err := r.queries.GetResultEventContext(ctx, db.GetResultEventContextParams{
		PEventID: eventID, PCompetitionID: competitionID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return EventContext{}, ErrEventNotFound
	}
	if err != nil {
		return EventContext{}, err
	}
	return EventContext{
		EventID: row.EventID, EventName: row.EventName, CompetitionID: row.CompetitionID,
		CompetitionName: row.CompetitionName, CompetitionStatus: row.CompetitionStatus,
		StartAt: row.StartAt.Time, EndAt: row.EndAt.Time,
		DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
		AgeGroupName: textPointer(row.AgeGroupName),
	}, nil
}

func (r *SQLCRepository) ListApprovedApplications(ctx context.Context, eventID int64) ([]Application, error) {
	rows, err := r.queries.ListApprovedResultApplications(ctx, eventID)
	if err != nil {
		return nil, err
	}
	result := make([]Application, 0, len(rows))
	for _, row := range rows {
		result = append(result, Application{
			ID: row.ApplicationID, AthleteUserID: row.AthleteUserID,
			LastName: row.LastName, FirstName: row.FirstName,
			MiddleName: textPointer(row.MiddleName),
		})
	}
	return result, nil
}

func (r *SQLCRepository) ListEligibleTeamMemberships(
	ctx context.Context, eventID int64,
) ([]TeamMembership, error) {
	rows, err := r.queries.ListEligibleResultTeamMemberships(ctx, eventID)
	if err != nil {
		return nil, err
	}
	result := make([]TeamMembership, 0, len(rows))
	for _, row := range rows {
		result = append(result, TeamMembership{
			ApplicationID: row.ApplicationID, AthleteUserID: row.AthleteUserID,
			TeamID: row.TeamID, TeamName: row.TeamName,
		})
	}
	return result, nil
}

func (r *SQLCRepository) Configure(
	ctx context.Context, eventID int64, mode string, taskCount int16, teams []TeamSetup, confirmReset bool,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)

	sheet, err := queries.LockEventResultSheet(ctx, eventID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		sheet, err = queries.CreateEventResultSheet(ctx, db.CreateEventResultSheetParams{
			PEventID: eventID, PMode: mode, PTaskCount: taskCount,
		})
	case err == nil && sheet.Status == SheetApproved:
		return ErrAlreadyApproved
	case err == nil:
		current, loadErr := loadTable(ctx, queries, eventID)
		if loadErr != nil {
			return loadErr
		}
		applications, loadErr := queries.ListApprovedResultApplications(ctx, eventID)
		if loadErr != nil {
			return loadErr
		}
		approved := make([]Application, 0, len(applications))
		for _, application := range applications {
			approved = append(approved, Application{ID: application.ApplicationID})
		}
		if sameConfiguration(current, mode, taskCount, teams, approved) {
			return tx.Commit(ctx)
		}
		if !confirmReset {
			return ErrResetConfirmation
		}
		sheet, err = queries.UpdateEventResultSheetSettings(ctx, db.UpdateEventResultSheetSettingsParams{
			PEventID: eventID, PMode: mode, PTaskCount: taskCount,
		})
	}
	if err != nil {
		return err
	}
	if err := queries.DeleteEventResultEntries(ctx, sheet.ID); err != nil {
		return err
	}

	if mode == ModeIndividual {
		applications, err := queries.ListApprovedResultApplications(ctx, eventID)
		if err != nil {
			return err
		}
		for _, application := range applications {
			entryID, err := queries.CreateEventResultEntry(ctx, db.CreateEventResultEntryParams{
				PSheetID: sheet.ID,
			})
			if err != nil {
				return err
			}
			if err := queries.LinkResultEntryApplication(ctx, db.LinkResultEntryApplicationParams{
				PEntryID: entryID, PApplicationID: application.ApplicationID,
			}); err != nil {
				return err
			}
		}
	} else {
		for _, team := range teams {
			entryID, err := queries.CreateEventResultEntry(ctx, db.CreateEventResultEntryParams{
				PSheetID: sheet.ID, PTeamID: pgtype.Int8{Int64: team.TeamID, Valid: true},
			})
			if err != nil {
				return err
			}
			for _, applicationID := range team.ApplicationIDs {
				if err := queries.LinkResultEntryApplication(ctx, db.LinkResultEntryApplicationParams{
					PEntryID: entryID, PApplicationID: applicationID,
				}); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit(ctx)
}

func (r *SQLCRepository) ConfigureManual(
	ctx context.Context, eventID int64, mode string, template ResultTemplate, teams []TeamSetup, confirmReset bool,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	snapshot, err := json.Marshal(TemplateSnapshot{ID: template.ID, Name: template.Name, Fields: template.Fields})
	if err != nil {
		return err
	}
	sheet, err := queries.LockEventResultSheet(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		err = tx.QueryRow(ctx, `INSERT INTO competition_event_result_sheet
			(event_id, mode, task_count, scoring_mode, template_id, template_snapshot)
			VALUES ($1,$2,0,'manual',$3,$4) RETURNING id`, eventID, mode, template.ID, snapshot).Scan(&sheet.ID)
		sheet.EventID, sheet.Mode, sheet.TaskCount, sheet.ScoringMode = eventID, mode, 0, ScoringManual
	} else if err == nil && sheet.Status == SheetApproved {
		return ErrAlreadyApproved
	} else if err == nil {
		current, loadErr := loadTable(ctx, queries, eventID)
		if loadErr != nil {
			return loadErr
		}
		applications, loadErr := queries.ListApprovedResultApplications(ctx, eventID)
		if loadErr != nil {
			return loadErr
		}
		approved := make([]Application, 0, len(applications))
		for _, a := range applications {
			approved = append(approved, Application{ID: a.ApplicationID})
		}
		if current.Sheet.ScoringMode == ScoringManual && current.Sheet.TemplateID != nil && *current.Sheet.TemplateID == template.ID && sameConfiguration(current, mode, 0, teams, approved) {
			return tx.Commit(ctx)
		}
		if !confirmReset {
			return ErrResetConfirmation
		}
		err = tx.QueryRow(ctx, `UPDATE competition_event_result_sheet SET mode=$2, task_count=0,
			scoring_mode='manual', template_id=$3, template_snapshot=$4, calculated_at=NULL,
			revision=revision+1, updated_at=now() WHERE event_id=$1 AND status='draft' RETURNING id`,
			eventID, mode, template.ID, snapshot).Scan(&sheet.ID)
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM competition_event_result_entry WHERE sheet_id=$1`, sheet.ID); err != nil {
		return err
	}
	if mode == ModeIndividual {
		applications, err := queries.ListApprovedResultApplications(ctx, eventID)
		if err != nil {
			return err
		}
		for _, application := range applications {
			entryID, err := queries.CreateEventResultEntry(ctx, db.CreateEventResultEntryParams{PSheetID: sheet.ID})
			if err != nil {
				return err
			}
			if err = queries.LinkResultEntryApplication(ctx, db.LinkResultEntryApplicationParams{PEntryID: entryID, PApplicationID: application.ApplicationID}); err != nil {
				return err
			}
		}
	} else {
		for _, team := range teams {
			entryID, err := queries.CreateEventResultEntry(ctx, db.CreateEventResultEntryParams{PSheetID: sheet.ID, PTeamID: pgtype.Int8{Int64: team.TeamID, Valid: true}})
			if err != nil {
				return err
			}
			for _, applicationID := range team.ApplicationIDs {
				if err = queries.LinkResultEntryApplication(ctx, db.LinkResultEntryApplicationParams{PEntryID: entryID, PApplicationID: applicationID}); err != nil {
					return err
				}
			}
		}
	}
	return tx.Commit(ctx)
}

func (r *SQLCRepository) GetTable(ctx context.Context, eventID int64) (Table, error) {
	return loadTable(ctx, r.queries, eventID)
}

func loadTable(ctx context.Context, queries *db.Queries, eventID int64) (Table, error) {
	sheetRow, err := queries.GetEventResultSheet(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Table{}, ErrNotConfigured
	}
	if err != nil {
		return Table{}, err
	}
	entryRows, err := queries.ListEventResultEntries(ctx, sheetRow.ID)
	if err != nil {
		return Table{}, err
	}
	applicationRows, err := queries.ListResultEntryApplications(ctx, sheetRow.ID)
	if err != nil {
		return Table{}, err
	}
	taskRows, err := queries.ListEventTaskResults(ctx, sheetRow.ID)
	if err != nil {
		return Table{}, err
	}

	entries := make([]Entry, 0, len(entryRows))
	entryIndexes := make(map[int64]int, len(entryRows))
	for _, row := range entryRows {
		entryIndexes[row.ID] = len(entries)
		entries = append(entries, Entry{
			ID: row.ID, TeamID: int64Pointer(row.TeamID), TeamName: textPointer(row.TeamName),
			Applications: make([]Application, 0), Tasks: make([]TaskResult, 0),
			ResultStatus: row.ResultStatus, SolvedCount: row.SolvedCount,
			TotalTimeDeciseconds:     row.TotalTimeDeciseconds,
			LastSolveTimeDeciseconds: row.LastSolveTimeDeciseconds,
			NegativeAttempts:         row.NegativeAttempts, Place: int32Pointer(row.Place),
			CustomValues: map[string]any{},
		})
		if len(row.CustomValues) > 0 {
			if err := json.Unmarshal(row.CustomValues, &entries[len(entries)-1].CustomValues); err != nil {
				return Table{}, err
			}
		}
	}
	for _, row := range applicationRows {
		index, ok := entryIndexes[row.EntryID]
		if !ok {
			continue
		}
		entries[index].Applications = append(entries[index].Applications, Application{
			ID: row.ApplicationID, AthleteUserID: row.AthleteUserID,
			LastName: row.LastName, FirstName: row.FirstName,
			MiddleName: textPointer(row.MiddleName),
		})
	}
	for _, row := range taskRows {
		index, ok := entryIndexes[row.EntryID]
		if !ok {
			continue
		}
		entries[index].Tasks = append(entries[index].Tasks, TaskResult{
			TaskNumber: row.TaskNumber, Status: row.Status,
			SolveTimeDeciseconds: int64Pointer(row.SolveTimeDeciseconds),
			NegativeAttempts:     row.NegativeAttempts,
		})
	}
	return Table{Sheet: mapSheet(sheetRow), Entries: entries}, nil
}

func (r *SQLCRepository) SaveManualEntries(ctx context.Context, eventID, revision int64, entries []ManualEntryUpdate) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var sheetID int64
	var status, scoring string
	var currentRevision int64
	err = tx.QueryRow(ctx, `SELECT id,status,scoring_mode,revision FROM competition_event_result_sheet WHERE event_id=$1 FOR UPDATE`, eventID).Scan(&sheetID, &status, &scoring, &currentRevision)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotConfigured
	}
	if err != nil {
		return err
	}
	if status == SheetApproved {
		return ErrAlreadyApproved
	}
	if scoring != ScoringManual {
		return ErrWrongScoringMode
	}
	if revision != currentRevision {
		return ErrStaleResults
	}
	for _, entry := range entries {
		raw, err := json.Marshal(entry.CustomValues)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `UPDATE competition_event_result_entry SET result_status=$1,place=$2,custom_values=$3,
			solved_count=0,total_time_deciseconds=0,last_solve_time_deciseconds=0,negative_attempts=0
			WHERE id=$4 AND sheet_id=$5`, entry.ResultStatus, nullableInt4(entry.Place), raw, entry.EntryID, sheetID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return ErrEntryNotFound
		}
	}
	_, err = tx.Exec(ctx, `UPDATE competition_event_result_sheet SET calculated_at=NULL,revision=revision+1,updated_at=now() WHERE id=$1`, sheetID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *SQLCRepository) SaveDraftEntries(
	ctx context.Context, eventID int64, entries []EntryUpdate,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	sheet, err := queries.LockEventResultSheet(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotConfigured
	}
	if err != nil {
		return err
	}
	if sheet.Status == SheetApproved {
		return ErrAlreadyApproved
	}
	for _, entry := range entries {
		rows, err := queries.SetResultEntryStatus(ctx, db.SetResultEntryStatusParams{
			PResultStatus: entry.ResultStatus, PEntryID: entry.EntryID, PSheetID: sheet.ID,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrEntryNotFound
		}
		for _, task := range entry.Tasks {
			if task.TaskNumber > sheet.TaskCount {
				return ErrStaleResults
			}
			if err := queries.UpsertEventTaskResult(ctx, db.UpsertEventTaskResultParams{
				PEntryID: entry.EntryID, PTaskNumber: task.TaskNumber, PStatus: task.Status,
				PSolveTimeDeciseconds: nullableInt8(task.SolveTimeDeciseconds),
				PNegativeAttempts:     task.NegativeAttempts,
			}); err != nil {
				return err
			}
		}
	}
	if err := queries.ClearEventResultSummaries(ctx, sheet.ID); err != nil {
		return err
	}
	if err := queries.MarkEventResultSheetDirty(ctx, sheet.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *SQLCRepository) SaveCalculated(
	ctx context.Context, eventID, version int64, summaries []EntrySummary,
) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	sheet, err := queries.LockEventResultSheet(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotConfigured
	}
	if err != nil {
		return err
	}
	if sheet.Status == SheetApproved {
		return ErrAlreadyApproved
	}
	if sheet.Revision != version {
		return ErrStaleResults
	}
	if err := updateSummaries(ctx, queries, sheet.ID, summaries); err != nil {
		return err
	}
	rows, err := queries.MarkEventResultSheetCalculated(ctx, sheet.ID)
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrAlreadyApproved
	}
	return tx.Commit(ctx)
}

func (r *SQLCRepository) Approve(
	ctx context.Context, eventID, actorUserID, version int64, summaries []EntrySummary,
	protocolKey, protocolURL string,
) (Approval, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Approval{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := r.queries.WithTx(tx)
	sheet, err := queries.LockEventResultSheet(ctx, eventID)
	if errors.Is(err, pgx.ErrNoRows) {
		return Approval{}, ErrNotConfigured
	}
	if err != nil {
		return Approval{}, err
	}
	if sheet.Status == SheetApproved {
		return Approval{ApprovedAt: sheet.ApprovedAt.Time, ProtocolURL: protocolURL}, nil
	}
	if sheet.Revision != version {
		return Approval{}, ErrStaleResults
	}
	current, err := loadTable(ctx, queries, eventID)
	if err != nil {
		return Approval{}, err
	}
	applications, err := queries.ListApprovedResultApplications(ctx, eventID)
	if err != nil {
		return Approval{}, err
	}
	approved := make([]Application, 0, len(applications))
	for _, application := range applications {
		approved = append(approved, Application{ID: application.ApplicationID})
	}
	if !sameApplications(current.Entries, approved) {
		return Approval{}, ErrApplicationsChanged
	}
	if err := updateSummaries(ctx, queries, sheet.ID, summaries); err != nil {
		return Approval{}, err
	}
	table, err := loadTable(ctx, queries, eventID)
	if err != nil {
		return Approval{}, err
	}
	summaryByEntry := make(map[int64]EntrySummary, len(summaries))
	for _, summary := range summaries {
		summaryByEntry[summary.EntryID] = summary
	}
	for _, entry := range table.Entries {
		summary, ok := summaryByEntry[entry.ID]
		if !ok {
			return Approval{}, ErrEntryNotFound
		}
		for _, application := range entry.Applications {
			score := pgtype.Numeric{}
			if sheet.ScoringMode == "" || sheet.ScoringMode == ScoringAlgorithmic {
				score = pgtype.Numeric{Int: big.NewInt(int64(summary.SolvedCount)), Valid: true}
			}
			_, err := queries.UpsertApprovedCompetitionResult(ctx, db.UpsertApprovedCompetitionResultParams{
				PApplicationID: application.ID,
				PPlace:         nullableInt4(summary.Place),
				PScore:         score,
				PResultStatus:  pgtype.Text{String: entry.ResultStatus, Valid: true},
			})
			if err != nil {
				return Approval{}, err
			}
		}
	}
	approvedAt, err := queries.ApproveEventResultSheet(ctx, db.ApproveEventResultSheetParams{
		PApprovedByUserID:   pgtype.Int8{Int64: actorUserID, Valid: true},
		PProtocolStorageKey: pgtype.Text{String: protocolKey, Valid: true},
		PSheetID:            sheet.ID,
	})
	if err != nil {
		return Approval{}, err
	}
	if err := queries.SetCompetitionEventProtocolURL(ctx, db.SetCompetitionEventProtocolURLParams{
		PProtocolUrl: pgtype.Text{String: protocolURL, Valid: true}, PEventID: eventID,
	}); err != nil {
		return Approval{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Approval{}, err
	}
	return Approval{ApprovedAt: approvedAt.Time, ProtocolURL: protocolURL}, nil
}

func updateSummaries(
	ctx context.Context, queries *db.Queries, sheetID int64, summaries []EntrySummary,
) error {
	for _, summary := range summaries {
		rows, err := queries.UpdateResultEntrySummary(ctx, db.UpdateResultEntrySummaryParams{
			PSolvedCount:              summary.SolvedCount,
			PTotalTimeDeciseconds:     summary.TotalTimeDeciseconds,
			PLastSolveTimeDeciseconds: summary.LastSolveTimeDeciseconds,
			PNegativeAttempts:         summary.NegativeAttempts,
			PPlace:                    nullableInt4(summary.Place),
			PEntryID:                  summary.EntryID, PSheetID: sheetID,
		})
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrEntryNotFound
		}
	}
	return nil
}

func nullableInt4(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func sameConfiguration(table Table, mode string, taskCount int16, teams []TeamSetup, approved []Application) bool {
	if table.Sheet.Mode != mode || table.Sheet.TaskCount != taskCount || !sameApplications(table.Entries, approved) {
		return false
	}
	if mode == ModeIndividual {
		for _, entry := range table.Entries {
			if entry.TeamID != nil || len(entry.Applications) != 1 {
				return false
			}
		}
		return true
	}
	if len(teams) != len(table.Entries) {
		return false
	}
	byTeam := make(map[int64]map[int64]bool, len(teams))
	for _, team := range teams {
		ids := make(map[int64]bool, len(team.ApplicationIDs))
		for _, id := range team.ApplicationIDs {
			ids[id] = true
		}
		byTeam[team.TeamID] = ids
	}
	for _, entry := range table.Entries {
		if entry.TeamID == nil {
			return false
		}
		ids, ok := byTeam[*entry.TeamID]
		if !ok || len(ids) != len(entry.Applications) {
			return false
		}
		for _, application := range entry.Applications {
			if !ids[application.ID] {
				return false
			}
		}
	}
	return true
}

func (r *SQLCRepository) ListProtocolJudges(ctx context.Context, competitionID int64) ([]Judge, error) {
	rows, err := r.queries.ListResultProtocolJudges(ctx, competitionID)
	if err != nil {
		return nil, err
	}
	result := make([]Judge, 0, len(rows))
	for _, row := range rows {
		result = append(result, Judge{
			PositionName: row.PositionName, LastName: row.LastName,
			FirstName: row.FirstName, MiddleName: textPointer(row.MiddleName),
		})
	}
	return result, nil
}

func (r *SQLCRepository) ListHistory(ctx context.Context, userID int64, disciplineID *int64) ([]HistoryItem, error) {
	rows, err := r.queries.ListAthleteResultHistory(ctx, db.ListAthleteResultHistoryParams{PUserID: userID, PDisciplineID: nullableInt8(disciplineID)})
	if err != nil {
		return nil, err
	}
	items := make([]HistoryItem, 0, len(rows))
	for _, row := range rows {
		var score *float64
		if row.Score.Valid {
			value, err := row.Score.Float64Value()
			if err != nil {
				return nil, err
			}
			score = &value.Float64
		}
		items = append(items, HistoryItem{ResultID: row.ResultID, CompetitionID: row.CompetitionID, CompetitionName: row.CompetitionName,
			EventID: row.EventID, DisciplineID: row.DisciplineID, DisciplineName: row.DisciplineName,
			StartAt: row.StartAt.Time, Place: int32Pointer(row.Place), Score: score,
			ResultStatus: textPointer(row.ResultStatus), ProtocolURL: textPointer(row.ProtocolUrl)})
	}
	return items, nil
}

func mapSheet(row db.CompetitionEventResultSheet) Sheet {
	result := Sheet{
		ID: row.ID, EventID: row.EventID, Mode: row.Mode, TaskCount: row.TaskCount,
		Status: row.Status, CalculatedAt: timePointer(row.CalculatedAt),
		ApprovedByUserID: int64Pointer(row.ApprovedByUserID), ApprovedAt: timePointer(row.ApprovedAt),
		ProtocolStorageKey: textPointer(row.ProtocolStorageKey),
		CreatedAt:          row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time, Revision: row.Revision,
		ScoringMode: row.ScoringMode, TemplateID: int64Pointer(row.TemplateID),
	}
	if len(row.TemplateSnapshot) > 0 {
		var snapshot TemplateSnapshot
		if json.Unmarshal(row.TemplateSnapshot, &snapshot) == nil {
			result.TemplateSnapshot = &snapshot
		}
	}
	return result
}

func textPointer(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	result := value.String
	return &result
}

func int64Pointer(value pgtype.Int8) *int64 {
	if !value.Valid {
		return nil
	}
	result := value.Int64
	return &result
}

func int32Pointer(value pgtype.Int4) *int32 {
	if !value.Valid {
		return nil
	}
	result := value.Int32
	return &result
}

func timePointer(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	result := value.Time
	return &result
}

func nullableInt8(value *int64) pgtype.Int8 {
	if value == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *value, Valid: true}
}
