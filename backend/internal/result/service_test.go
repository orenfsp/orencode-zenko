package result

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

type repositoryStub struct {
	contextData     EventContext
	applications    []Application
	memberships     []TeamMembership
	table           Table
	err             error
	configuredMode  string
	configuredTeams []TeamSetup
	saved           []EntryUpdate
	summaries       []EntrySummary
	approval        Approval
	approveCalls    int
}

func (r *repositoryStub) GetEventContext(context.Context, int64, int64) (EventContext, error) {
	if r.err != nil {
		return EventContext{}, r.err
	}
	return r.contextData, nil
}
func (r *repositoryStub) ListApprovedApplications(context.Context, int64) ([]Application, error) {
	return r.applications, r.err
}
func (r *repositoryStub) ListEligibleTeamMemberships(context.Context, int64) ([]TeamMembership, error) {
	return r.memberships, r.err
}
func (r *repositoryStub) Configure(_ context.Context, eventID int64, mode string, taskCount int16, teams []TeamSetup, _ bool) error {
	if r.err != nil {
		return r.err
	}
	r.configuredMode, r.configuredTeams = mode, teams
	r.table.Sheet = Sheet{ID: 1, EventID: eventID, Mode: mode, TaskCount: taskCount, Status: SheetDraft}
	return nil
}
func (r *repositoryStub) GetTable(context.Context, int64) (Table, error) {
	if r.err != nil {
		return Table{}, r.err
	}
	return r.table, nil
}
func (r *repositoryStub) SaveDraftEntries(_ context.Context, _ int64, entries []EntryUpdate) error {
	if r.err != nil {
		return r.err
	}
	r.saved = entries
	return nil
}
func (r *repositoryStub) SaveCalculated(_ context.Context, _ int64, _ int64, summaries []EntrySummary) error {
	if r.err != nil {
		return r.err
	}
	r.summaries = summaries
	applySummaries(&r.table, summaries)
	now := time.Now()
	r.table.Sheet.CalculatedAt = &now
	return nil
}
func (r *repositoryStub) Approve(_ context.Context, _ int64, _ int64, _ int64, summaries []EntrySummary, _, protocolURL string) (Approval, error) {
	r.approveCalls++
	r.summaries = summaries
	if r.err != nil {
		return Approval{}, r.err
	}
	if r.approval.ApprovedAt.IsZero() {
		r.approval = Approval{ApprovedAt: time.Now(), ProtocolURL: protocolURL}
	}
	return r.approval, nil
}
func (r *repositoryStub) ListProtocolJudges(context.Context, int64) ([]Judge, error) {
	return []Judge{{PositionName: "главный судья", LastName: "Иванов", FirstName: "Иван"}}, r.err
}
func (r *repositoryStub) ListHistory(context.Context, int64, *int64) ([]HistoryItem, error) {
	return nil, r.err
}

type storageStub struct {
	data     []byte
	putCalls int
}

func (s *storageStub) Put(_ context.Context, _ string, src io.Reader) (int64, error) {
	s.putCalls++
	content, err := io.ReadAll(src)
	s.data = content
	return int64(len(content)), err
}
func (s *storageStub) Open(context.Context, string) (io.ReadCloser, error) {
	if len(s.data) == 0 {
		return nil, errors.New("not found")
	}
	return io.NopCloser(bytes.NewReader(s.data)), nil
}
func (*storageStub) Delete(context.Context, string) error { return nil }

type protocolStub struct {
	content []byte
	calls   int
}

func (p *protocolStub) Generate(ProtocolData) ([]byte, error) { p.calls++; return p.content, nil }

func algorithmicContext() EventContext {
	return EventContext{EventID: 2, CompetitionID: 1, CompetitionName: "Кубок", DisciplineID: 7, DisciplineName: AlgorithmicDiscipline}
}

func draftTable() Table {
	timeValue := int64(123)
	return Table{Sheet: Sheet{ID: 4, EventID: 2, Mode: ModeIndividual, TaskCount: 2, Status: SheetDraft}, Entries: []Entry{{
		ID: 9, ResultStatus: ResultFinished,
		Applications: []Application{{ID: 5, AthleteUserID: 20, LastName: "Иванов", FirstName: "Иван"}},
		Tasks:        []TaskResult{solvedTask(1, timeValue, 1)},
	}}}
}

func TestServiceConfiguresIndividualEntriesFromApprovedApplications(t *testing.T) {
	repo := &repositoryStub{contextData: algorithmicContext(), applications: []Application{{ID: 5}}, table: draftTable()}
	service := NewService(repo, &storageStub{}, &protocolStub{})
	response, err := service.Configure(context.Background(), 1, 2, SettingsRequest{Mode: ModeIndividual, TaskCount: 12})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	if repo.configuredMode != ModeIndividual || response.TaskCount != 12 || response.Entries == nil {
		t.Fatalf("configuration = %#v response = %#v", repo, response)
	}
}

func TestServiceValidatesTeamCompositionAgainstApprovedApplications(t *testing.T) {
	apps := []Application{{ID: 5}, {ID: 6}}
	memberships := []TeamMembership{{ApplicationID: 5, TeamID: 3}, {ApplicationID: 6, TeamID: 3}}
	repo := &repositoryStub{contextData: algorithmicContext(), applications: apps, memberships: memberships, table: draftTable()}
	service := NewService(repo, &storageStub{}, &protocolStub{})
	_, err := service.Configure(context.Background(), 1, 2, SettingsRequest{
		Mode: ModeTeam, TaskCount: 2,
		Teams: []TeamSetupRequest{{TeamID: 3, ApplicationIDs: []int64{5, 6}}},
	})
	if err != nil || len(repo.configuredTeams) != 1 {
		t.Fatalf("Configure(team) = %v, %#v", err, repo.configuredTeams)
	}
	_, err = service.Configure(context.Background(), 1, 2, SettingsRequest{
		Mode: ModeTeam, TaskCount: 2,
		Teams: []TeamSetupRequest{{TeamID: 3, ApplicationIDs: []int64{5}}},
	})
	if !errors.Is(err, ErrInvalidTeamSetup) {
		t.Fatalf("incomplete team error = %v", err)
	}
}

func TestServiceSavesDraftTaskResultsAndRejectsApprovedChanges(t *testing.T) {
	repo := &repositoryStub{contextData: algorithmicContext(), table: draftTable()}
	service := NewService(repo, &storageStub{}, &protocolStub{})
	timeValue := int64(321)
	_, err := service.Save(context.Background(), 1, 2, SaveResultsRequest{Entries: []EntryRequest{{
		EntryID: 9, ResultStatus: ResultFinished,
		Tasks: []TaskRequest{{TaskNumber: 1, Status: TaskSolved, SolveTimeDeciseconds: &timeValue, NegativeAttempts: 2}},
	}}})
	if err != nil || len(repo.saved) != 1 || repo.saved[0].Tasks[0].NegativeAttempts != 2 {
		t.Fatalf("Save() = %v, %#v", err, repo.saved)
	}
	repo.table.Sheet.Status = SheetApproved
	if _, err := service.Save(context.Background(), 1, 2, SaveResultsRequest{Entries: []EntryRequest{{EntryID: 9}}}); !errors.Is(err, ErrAlreadyApproved) {
		t.Fatalf("approved Save() error = %v", err)
	}
}

func TestServiceCalculatesAndRejectsUnsupportedDiscipline(t *testing.T) {
	repo := &repositoryStub{contextData: algorithmicContext(), table: draftTable()}
	service := NewService(repo, &storageStub{}, &protocolStub{})
	response, err := service.Calculate(context.Background(), 1, 2)
	if err != nil || len(repo.summaries) != 1 || repo.summaries[0].SolvedCount != 1 || response.Entries[0].Place == nil {
		t.Fatalf("Calculate() = %v, %#v", err, repo.summaries)
	}
	repo.contextData.DisciplineName = "программирование робототехники"
	if _, err := service.Calculate(context.Background(), 1, 2); !errors.Is(err, ErrUnsupportedDiscipline) {
		t.Fatalf("unsupported Calculate() error = %v", err)
	}
}

func TestServiceApprovesOnceAndStoresProtocol(t *testing.T) {
	repo := &repositoryStub{
		contextData: algorithmicContext(), table: draftTable(),
		applications: []Application{{ID: 5, AthleteUserID: 20}},
	}
	storage := &storageStub{}
	protocol := &protocolStub{content: []byte("%PDF-test")}
	service := NewService(repo, storage, protocol)
	response, err := service.Approve(context.Background(), 1, 2, 42)
	if err != nil || response.Status != SheetApproved || storage.putCalls != 1 || repo.approveCalls != 1 {
		t.Fatalf("Approve() = %#v, %v, puts=%d calls=%d", response, err, storage.putCalls, repo.approveCalls)
	}
	repo.table.Sheet.Status = SheetApproved
	repo.table.Sheet.ApprovedAt = &response.ApprovedAt
	response2, err := service.Approve(context.Background(), 1, 2, 42)
	if err != nil || response2.ProtocolURL != response.ProtocolURL || storage.putCalls != 1 || repo.approveCalls != 1 {
		t.Fatalf("idempotent Approve() = %#v, %v", response2, err)
	}
}

func TestServiceRejectsApprovalWhenApprovedApplicationsChanged(t *testing.T) {
	repo := &repositoryStub{
		contextData: algorithmicContext(), table: draftTable(),
		applications: []Application{{ID: 5}, {ID: 6}},
	}
	service := NewService(repo, &storageStub{}, &protocolStub{content: []byte("%PDF-test")})
	if _, err := service.Approve(context.Background(), 1, 2, 42); !errors.Is(err, ErrApplicationsChanged) {
		t.Fatalf("Approve() error = %v", err)
	}
}

func TestServicePropagatesRepositoryError(t *testing.T) {
	want := errors.New("database unavailable")
	service := NewService(&repositoryStub{err: want}, &storageStub{}, &protocolStub{})
	if _, err := service.Get(context.Background(), 1, 2); !errors.Is(err, want) {
		t.Fatalf("Get() error = %v", err)
	}
}
