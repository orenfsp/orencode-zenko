package result

import (
	"bytes"
	"context"
	"fmt"
	"github.com/google/uuid"
	"io"
	"math"
	"strings"
	"time"

	"github.com/orenfsp/orencode-zenko.git/pkg/logger"
	"go.uber.org/zap"
)

type Storage interface {
	Put(ctx context.Context, key string, src io.Reader) (int64, error)
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

type ProtocolGenerator interface {
	Generate(data ProtocolData) ([]byte, error)
}

type Service struct {
	repository Repository
	storage    Storage
	protocol   ProtocolGenerator
}

func NewService(repository Repository, storage Storage, protocol ProtocolGenerator) *Service {
	return &Service{repository: repository, storage: storage, protocol: protocol}
}

func (s *Service) Configure(
	ctx context.Context, competitionID, eventID int64, request SettingsRequest,
) (TableResponse, error) {
	contextData, err := s.getEvent(ctx, competitionID, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	if request.ScoringMode == "" {
		request.ScoringMode = ScoringAlgorithmic
	}
	if request.Mode != ModeIndividual && request.Mode != ModeTeam {
		return TableResponse{}, ErrInvalidRequest
	}
	if request.ScoringMode == ScoringAlgorithmic {
		if !strings.EqualFold(strings.TrimSpace(contextData.DisciplineName), AlgorithmicDiscipline) {
			return TableResponse{}, ErrUnsupportedDiscipline
		}
		if request.TaskCount < 1 || request.TaskCount > MaxTaskCount || request.TemplateID != nil {
			return TableResponse{}, ErrInvalidRequest
		}
	} else if request.ScoringMode == ScoringManual {
		if request.TaskCount != 0 || request.TemplateID == nil || *request.TemplateID <= 0 {
			return TableResponse{}, ErrInvalidRequest
		}
	} else {
		return TableResponse{}, ErrInvalidRequest
	}
	applications, err := s.repository.ListApprovedApplications(ctx, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	teams := make([]TeamSetup, 0, len(request.Teams))
	if request.Mode == ModeIndividual {
		if len(request.Teams) != 0 {
			return TableResponse{}, ErrInvalidTeamSetup
		}
	} else {
		memberships, err := s.repository.ListEligibleTeamMemberships(ctx, eventID)
		if err != nil {
			return TableResponse{}, err
		}
		teams, err = validateTeamSetup(applications, memberships, request.Teams)
		if err != nil {
			return TableResponse{}, err
		}
	}
	if request.ScoringMode == ScoringAlgorithmic {
		err = s.repository.Configure(ctx, eventID, request.Mode, request.TaskCount, teams, request.ConfirmReset)
	} else {
		template, loadErr := s.repository.GetTemplate(ctx, *request.TemplateID)
		if loadErr != nil {
			return TableResponse{}, loadErr
		}
		if !template.IsActive {
			return TableResponse{}, ErrTemplateInactive
		}
		if template.DisciplineID != contextData.DisciplineID {
			return TableResponse{}, ErrTemplateDiscipline
		}
		err = s.repository.ConfigureManual(ctx, eventID, request.Mode, template, teams, request.ConfirmReset)
	}
	if err != nil {
		return TableResponse{}, err
	}
	table, err := s.repository.GetTable(ctx, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	table.Context = contextData
	return toTableResponse(table), nil
}

func (s *Service) Get(
	ctx context.Context, competitionID, eventID int64,
) (TableResponse, error) {
	contextData, err := s.getEvent(ctx, competitionID, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	table, err := s.repository.GetTable(ctx, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	table.Context = contextData
	return toTableResponse(table), nil
}

func (s *Service) Save(
	ctx context.Context, competitionID, eventID int64, request SaveResultsRequest,
) (TableResponse, error) {
	if _, err := s.requireAlgorithmicEvent(ctx, competitionID, eventID); err != nil {
		return TableResponse{}, err
	}
	table, err := s.repository.GetTable(ctx, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	if table.Sheet.Status == SheetApproved {
		return TableResponse{}, ErrAlreadyApproved
	}
	if effectiveScoringMode(table.Sheet) != ScoringAlgorithmic {
		return TableResponse{}, ErrWrongScoringMode
	}
	updates, err := validateEntryUpdates(table, request)
	if err != nil {
		return TableResponse{}, err
	}
	if err := s.repository.SaveDraftEntries(ctx, eventID, updates); err != nil {
		return TableResponse{}, err
	}
	return s.Get(ctx, competitionID, eventID)
}

func (s *Service) SaveManual(ctx context.Context, competitionID, eventID int64, request SaveManualResultsRequest) (TableResponse, error) {
	if _, err := s.getEvent(ctx, competitionID, eventID); err != nil {
		return TableResponse{}, err
	}
	table, err := s.repository.GetTable(ctx, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	if table.Sheet.Status == SheetApproved {
		return TableResponse{}, ErrAlreadyApproved
	}
	if effectiveScoringMode(table.Sheet) != ScoringManual || table.Sheet.TemplateSnapshot == nil {
		return TableResponse{}, ErrWrongScoringMode
	}
	updates, err := validateManualUpdates(table, request, true)
	if err != nil {
		return TableResponse{}, err
	}
	if err = s.repository.SaveManualEntries(ctx, eventID, request.Revision, updates); err != nil {
		return TableResponse{}, err
	}
	return s.Get(ctx, competitionID, eventID)
}

func (s *Service) Calculate(
	ctx context.Context, competitionID, eventID int64,
) (TableResponse, error) {
	contextData, err := s.getEvent(ctx, competitionID, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	table, err := s.repository.GetTable(ctx, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	if table.Sheet.Status == SheetApproved {
		return TableResponse{}, ErrAlreadyApproved
	}
	var summaries []EntrySummary
	if effectiveScoringMode(table.Sheet) == ScoringAlgorithmic {
		if !strings.EqualFold(strings.TrimSpace(contextData.DisciplineName), AlgorithmicDiscipline) {
			return TableResponse{}, ErrUnsupportedDiscipline
		}
		summaries = calculateAlgorithmic(table.Entries)
	} else if effectiveScoringMode(table.Sheet) == ScoringManual {
		if _, err = validateManualTable(table); err != nil {
			return TableResponse{}, err
		}
		summaries = manualSummaries(table.Entries)
	} else {
		return TableResponse{}, ErrWrongScoringMode
	}
	if err := s.repository.SaveCalculated(ctx, eventID, table.Sheet.Revision, summaries); err != nil {
		return TableResponse{}, err
	}
	table, err = s.repository.GetTable(ctx, eventID)
	if err != nil {
		return TableResponse{}, err
	}
	table.Context = contextData
	return toTableResponse(table), nil
}

func (s *Service) Approve(
	ctx context.Context, competitionID, eventID, actorUserID int64,
) (ApprovalResponse, error) {
	if actorUserID <= 0 {
		return ApprovalResponse{}, ErrInvalidRequest
	}
	contextData, err := s.getEvent(ctx, competitionID, eventID)
	if err != nil {
		return ApprovalResponse{}, err
	}
	table, err := s.repository.GetTable(ctx, eventID)
	if err != nil {
		return ApprovalResponse{}, err
	}
	protocolURL := protocolURL(competitionID, eventID)
	if table.Sheet.Status == SheetApproved {
		if table.Sheet.ApprovedAt == nil || table.Sheet.ProtocolStorageKey == nil {
			return ApprovalResponse{}, ErrProtocolNotFound
		}
		return ApprovalResponse{
			Status: SheetApproved, ApprovedAt: *table.Sheet.ApprovedAt, ProtocolURL: protocolURL,
		}, nil
	}
	applications, err := s.repository.ListApprovedApplications(ctx, eventID)
	if err != nil {
		return ApprovalResponse{}, err
	}
	if !sameApplications(table.Entries, applications) {
		return ApprovalResponse{}, ErrApplicationsChanged
	}
	if len(table.Entries) == 0 {
		return ApprovalResponse{}, ErrInvalidRequest
	}
	var summaries []EntrySummary
	if effectiveScoringMode(table.Sheet) == ScoringAlgorithmic {
		if !strings.EqualFold(strings.TrimSpace(contextData.DisciplineName), AlgorithmicDiscipline) {
			return ApprovalResponse{}, ErrUnsupportedDiscipline
		}
		summaries = calculateAlgorithmic(table.Entries)
	} else if effectiveScoringMode(table.Sheet) == ScoringManual {
		if _, err = validateManualTable(table); err != nil {
			return ApprovalResponse{}, err
		}
		summaries = manualSummaries(table.Entries)
	} else {
		return ApprovalResponse{}, ErrWrongScoringMode
	}
	applySummaries(&table, summaries)
	judges, err := s.repository.ListProtocolJudges(ctx, competitionID)
	if err != nil {
		return ApprovalResponse{}, err
	}
	table.Context = contextData
	table.Judges = judges
	generatedAt := time.Now().UTC()
	content, err := s.protocol.Generate(ProtocolData{Table: table, Date: generatedAt})
	if err != nil {
		return ApprovalResponse{}, err
	}
	key := fmt.Sprintf("protocols/competition-%d/event-%d-%s.pdf", competitionID, eventID, uuid.NewString())
	if _, err := s.storage.Put(ctx, key, bytes.NewReader(content)); err != nil {
		return ApprovalResponse{}, err
	}
	approval, err := s.repository.Approve(
		ctx, eventID, actorUserID, table.Sheet.Revision, summaries, key, protocolURL,
	)
	if err != nil {
		return ApprovalResponse{}, err
	}
	return ApprovalResponse{
		Status: SheetApproved, ApprovedAt: approval.ApprovedAt, ProtocolURL: approval.ProtocolURL,
	}, nil
}

func (s *Service) UploadProtocol(ctx context.Context, competitionID, eventID int64, content []byte) (string, error) {
	if _, err := s.getEvent(ctx, competitionID, eventID); err != nil {
		return "", err
	}
	key := fmt.Sprintf("protocols/competition-%d/event-%d-%s.pdf", competitionID, eventID, uuid.NewString())
	if _, err := s.storage.Put(ctx, key, bytes.NewReader(content)); err != nil {
		return "", err
	}
	url := protocolURL(competitionID, eventID)
	oldKey, err := s.repository.ReplaceProtocol(ctx, eventID, &key, &url)
	if err != nil {
		s.removeProtocolFile(key)
		return "", err
	}
	if oldKey != nil {
		s.removeProtocolFile(*oldKey)
	}
	return url, nil
}

func (s *Service) DeleteProtocol(ctx context.Context, competitionID, eventID int64) error {
	if _, err := s.getEvent(ctx, competitionID, eventID); err != nil {
		return err
	}
	oldKey, err := s.repository.ReplaceProtocol(ctx, eventID, nil, nil)
	if err != nil {
		return err
	}
	if oldKey != nil {
		s.removeProtocolFile(*oldKey)
	}
	return nil
}

func (s *Service) removeProtocolFile(key string) {
	if err := s.storage.Delete(context.Background(), key); err != nil && logger.Log != nil {
		logger.Log.Error("remove obsolete protocol file", zap.Error(err))
	}
}

func sameApplications(entries []Entry, applications []Application) bool {
	linked := make(map[int64]struct{}, len(applications))
	for _, entry := range entries {
		for _, application := range entry.Applications {
			if _, exists := linked[application.ID]; exists {
				return false
			}
			linked[application.ID] = struct{}{}
		}
	}
	if len(linked) != len(applications) {
		return false
	}
	for _, application := range applications {
		if _, exists := linked[application.ID]; !exists {
			return false
		}
	}
	return true
}

func (s *Service) OpenProtocol(
	ctx context.Context, competitionID, eventID int64,
) (ProtocolDownload, error) {
	if _, err := s.getEvent(ctx, competitionID, eventID); err != nil {
		return ProtocolDownload{}, err
	}
	table, err := s.repository.GetTable(ctx, eventID)
	if err != nil {
		return ProtocolDownload{}, err
	}
	if table.Sheet.Status != SheetApproved || table.Sheet.ProtocolStorageKey == nil {
		return ProtocolDownload{}, ErrProtocolNotFound
	}
	content, err := s.storage.Open(ctx, *table.Sheet.ProtocolStorageKey)
	if err != nil {
		return ProtocolDownload{}, ErrProtocolNotFound
	}
	return ProtocolDownload{
		Content: content, Filename: fmt.Sprintf("competition-%d-event-%d-protocol.pdf", competitionID, eventID),
	}, nil
}

func (s *Service) getEvent(ctx context.Context, competitionID, eventID int64) (EventContext, error) {
	if competitionID <= 0 || eventID <= 0 {
		return EventContext{}, ErrInvalidRequest
	}
	return s.repository.GetEventContext(ctx, competitionID, eventID)
}

func (s *Service) requireAlgorithmicEvent(
	ctx context.Context, competitionID, eventID int64,
) (EventContext, error) {
	contextData, err := s.getEvent(ctx, competitionID, eventID)
	if err != nil {
		return EventContext{}, err
	}
	if !strings.EqualFold(strings.TrimSpace(contextData.DisciplineName), AlgorithmicDiscipline) {
		return EventContext{}, ErrUnsupportedDiscipline
	}
	return contextData, nil
}

func validateTeamSetup(
	applications []Application, memberships []TeamMembership, requests []TeamSetupRequest,
) ([]TeamSetup, error) {
	approved := make(map[int64]struct{}, len(applications))
	for _, application := range applications {
		approved[application.ID] = struct{}{}
	}
	eligible := make(map[[2]int64]struct{}, len(memberships))
	for _, membership := range memberships {
		eligible[[2]int64{membership.ApplicationID, membership.TeamID}] = struct{}{}
	}
	seenApplications := make(map[int64]struct{}, len(applications))
	seenTeams := make(map[int64]struct{}, len(requests))
	result := make([]TeamSetup, 0, len(requests))
	for _, request := range requests {
		if request.TeamID <= 0 || len(request.ApplicationIDs) == 0 {
			return nil, ErrInvalidTeamSetup
		}
		if _, exists := seenTeams[request.TeamID]; exists {
			return nil, ErrInvalidTeamSetup
		}
		seenTeams[request.TeamID] = struct{}{}
		ids := make([]int64, 0, len(request.ApplicationIDs))
		for _, applicationID := range request.ApplicationIDs {
			if _, ok := approved[applicationID]; !ok {
				return nil, ErrInvalidTeamSetup
			}
			if _, ok := eligible[[2]int64{applicationID, request.TeamID}]; !ok {
				return nil, ErrInvalidTeamSetup
			}
			if _, exists := seenApplications[applicationID]; exists {
				return nil, ErrInvalidTeamSetup
			}
			seenApplications[applicationID] = struct{}{}
			ids = append(ids, applicationID)
		}
		result = append(result, TeamSetup{TeamID: request.TeamID, ApplicationIDs: ids})
	}
	if len(seenApplications) != len(approved) {
		return nil, ErrInvalidTeamSetup
	}
	return result, nil
}

func validateEntryUpdates(table Table, request SaveResultsRequest) ([]EntryUpdate, error) {
	if len(request.Entries) == 0 {
		return nil, ErrInvalidRequest
	}
	validEntries := make(map[int64]struct{}, len(table.Entries))
	for _, entry := range table.Entries {
		validEntries[entry.ID] = struct{}{}
	}
	seenEntries := make(map[int64]struct{}, len(request.Entries))
	result := make([]EntryUpdate, 0, len(request.Entries))
	for _, entry := range request.Entries {
		if _, ok := validEntries[entry.EntryID]; !ok {
			return nil, ErrEntryNotFound
		}
		if _, exists := seenEntries[entry.EntryID]; exists {
			return nil, ErrInvalidRequest
		}
		seenEntries[entry.EntryID] = struct{}{}
		if entry.ResultStatus != ResultFinished && entry.ResultStatus != ResultDidNotStart &&
			entry.ResultStatus != ResultDisqualified {
			return nil, ErrInvalidRequest
		}
		if entry.ResultStatus != ResultFinished && len(entry.Tasks) != 0 {
			return nil, ErrInvalidRequest
		}
		seenTasks := make(map[int16]struct{}, len(entry.Tasks))
		tasks := make([]TaskResult, 0, len(entry.Tasks))
		for _, task := range entry.Tasks {
			if task.TaskNumber < 1 || task.TaskNumber > table.Sheet.TaskCount || task.NegativeAttempts < 0 {
				return nil, ErrInvalidRequest
			}
			if _, exists := seenTasks[task.TaskNumber]; exists {
				return nil, ErrInvalidRequest
			}
			seenTasks[task.TaskNumber] = struct{}{}
			switch task.Status {
			case TaskSolved:
				if task.SolveTimeDeciseconds == nil || *task.SolveTimeDeciseconds < 0 {
					return nil, ErrInvalidRequest
				}
			case TaskAttempted:
				if task.SolveTimeDeciseconds != nil {
					return nil, ErrInvalidRequest
				}
			case TaskNotAttempted:
				if task.SolveTimeDeciseconds != nil || task.NegativeAttempts != 0 {
					return nil, ErrInvalidRequest
				}
			default:
				return nil, ErrInvalidRequest
			}
			tasks = append(tasks, TaskResult{
				TaskNumber: task.TaskNumber, Status: task.Status,
				SolveTimeDeciseconds: task.SolveTimeDeciseconds,
				NegativeAttempts:     task.NegativeAttempts,
			})
		}
		result = append(result, EntryUpdate{
			EntryID: entry.EntryID, ResultStatus: entry.ResultStatus, Tasks: tasks,
		})
	}
	return result, nil
}

func applySummaries(table *Table, summaries []EntrySummary) {
	byID := make(map[int64]EntrySummary, len(summaries))
	for _, summary := range summaries {
		byID[summary.EntryID] = summary
	}
	for index := range table.Entries {
		summary := byID[table.Entries[index].ID]
		table.Entries[index].SolvedCount = summary.SolvedCount
		table.Entries[index].TotalTimeDeciseconds = summary.TotalTimeDeciseconds
		table.Entries[index].LastSolveTimeDeciseconds = summary.LastSolveTimeDeciseconds
		table.Entries[index].NegativeAttempts = summary.NegativeAttempts
		table.Entries[index].Place = summary.Place
	}
}

func protocolURL(competitionID, eventID int64) string {
	return fmt.Sprintf("/api/v1/competitions/%d/events/%d/protocol", competitionID, eventID)
}

func toTableResponse(table Table) TableResponse {
	entries := make([]EntryResponse, 0, len(table.Entries))
	for _, entry := range table.Entries {
		participants := make([]ParticipantResponse, 0, len(entry.Applications))
		for _, application := range entry.Applications {
			participants = append(participants, ParticipantResponse{
				ApplicationID: application.ID, UserID: application.AthleteUserID,
				LastName: application.LastName, FirstName: application.FirstName,
				MiddleName: application.MiddleName,
			})
		}
		tasks := make([]TaskResponse, 0, len(entry.Tasks))
		for _, task := range entry.Tasks {
			tasks = append(tasks, TaskResponse{
				TaskNumber: task.TaskNumber, Status: task.Status,
				SolveTimeDeciseconds: task.SolveTimeDeciseconds,
				NegativeAttempts:     task.NegativeAttempts,
			})
		}
		var team *TeamResponse
		if entry.TeamID != nil && entry.TeamName != nil {
			team = &TeamResponse{ID: *entry.TeamID, Name: *entry.TeamName}
		}
		entries = append(entries, EntryResponse{
			ID: entry.ID, Team: team, Participants: participants,
			ResultStatus: entry.ResultStatus, Tasks: tasks,
			SolvedCount: entry.SolvedCount, TotalTimeDeciseconds: entry.TotalTimeDeciseconds,
			LastSolveTimeDeciseconds: entry.LastSolveTimeDeciseconds,
			NegativeAttempts:         entry.NegativeAttempts, Place: entry.Place,
			CustomValues: entry.CustomValues,
		})
	}
	var protocol *string
	if table.Sheet.Status == SheetApproved && table.Sheet.ProtocolStorageKey != nil {
		value := protocolURL(table.Context.CompetitionID, table.Context.EventID)
		protocol = &value
	}
	return TableResponse{
		Competition: CompetitionResponse{ID: table.Context.CompetitionID, Name: table.Context.CompetitionName},
		Event: EventResponse{ID: table.Context.EventID, Name: table.Context.EventName, DisciplineID: table.Context.DisciplineID,
			DisciplineName: table.Context.DisciplineName, AgeGroupName: table.Context.AgeGroupName},
		Mode: table.Sheet.Mode, TaskCount: table.Sheet.TaskCount, Status: table.Sheet.Status,
		CalculatedAt: table.Sheet.CalculatedAt, ApprovedAt: table.Sheet.ApprovedAt,
		ProtocolURL: protocol, Entries: entries,
		Revision: table.Sheet.Revision, ScoringMode: table.Sheet.ScoringMode,
		TemplateID: table.Sheet.TemplateID, Template: table.Sheet.TemplateSnapshot,
	}
}

func validateManualUpdates(table Table, request SaveManualResultsRequest, requireAll bool) ([]ManualEntryUpdate, error) {
	if table.Sheet.TemplateSnapshot == nil || len(request.Entries) == 0 {
		return nil, ErrInvalidRequest
	}
	valid := make(map[int64]bool, len(table.Entries))
	for _, e := range table.Entries {
		valid[e.ID] = true
	}
	seen, places := map[int64]bool{}, map[int32]bool{}
	updates := make([]ManualEntryUpdate, 0, len(request.Entries))
	for _, item := range request.Entries {
		if !valid[item.EntryID] {
			return nil, ErrEntryNotFound
		}
		if seen[item.EntryID] {
			return nil, ErrInvalidRequest
		}
		seen[item.EntryID] = true
		if item.ResultStatus != ResultFinished && item.ResultStatus != ResultDidNotStart && item.ResultStatus != ResultDisqualified {
			return nil, ErrInvalidRequest
		}
		if item.ResultStatus == ResultFinished {
			if item.Place == nil || *item.Place <= 0 || places[*item.Place] {
				return nil, ErrInvalidRequest
			}
			places[*item.Place] = true
		} else if item.Place != nil {
			return nil, ErrInvalidRequest
		}
		if err := validateCustomValues(table.Sheet.TemplateSnapshot.Fields, item.CustomValues, item.ResultStatus == ResultFinished); err != nil {
			return nil, err
		}
		updates = append(updates, ManualEntryUpdate{EntryID: item.EntryID, ResultStatus: item.ResultStatus, Place: item.Place, CustomValues: item.CustomValues})
	}
	if requireAll && len(seen) != len(valid) {
		return nil, ErrInvalidRequest
	}
	return updates, nil
}

func validateCustomValues(fields []TemplateField, values map[string]any, required bool) error {
	allowed := make(map[string]TemplateField, len(fields))
	for _, f := range fields {
		allowed[f.Key] = f
	}
	for key, value := range values {
		field, ok := allowed[key]
		if !ok {
			return ErrInvalidRequest
		}
		switch field.Type {
		case "text":
			v, ok := value.(string)
			if !ok || len([]rune(v)) > MaxManualTextLength {
				return ErrInvalidRequest
			}
		case "number":
			v, ok := value.(float64)
			if !ok || math.IsNaN(v) || math.IsInf(v, 0) {
				return ErrInvalidRequest
			}
		default:
			return ErrInvalidRequest
		}
	}
	if required {
		for _, f := range fields {
			if f.Required {
				if v, ok := values[f.Key]; !ok || v == nil || (f.Type == "text" && strings.TrimSpace(v.(string)) == "") {
					return ErrInvalidRequest
				}
			}
		}
	}
	return nil
}

func validateManualTable(table Table) ([]ManualEntryUpdate, error) {
	request := SaveManualResultsRequest{Entries: make([]ManualEntryRequest, 0, len(table.Entries))}
	for _, e := range table.Entries {
		request.Entries = append(request.Entries, ManualEntryRequest{EntryID: e.ID, ResultStatus: e.ResultStatus, Place: e.Place, CustomValues: e.CustomValues})
	}
	return validateManualUpdates(table, request, true)
}

func manualSummaries(entries []Entry) []EntrySummary {
	items := make([]EntrySummary, 0, len(entries))
	for _, e := range entries {
		items = append(items, EntrySummary{EntryID: e.ID, Place: e.Place})
	}
	return items
}

func effectiveScoringMode(sheet Sheet) string {
	if sheet.ScoringMode == "" {
		return ScoringAlgorithmic
	}
	return sheet.ScoringMode
}
