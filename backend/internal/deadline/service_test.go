package deadline

import (
	"context"
	"testing"
	"time"
)

type repositoryStub struct {
	roles          []string
	registration   []Item
	athleteRanks   []Item
	coachRanks     []Item
	allRanks       []Item
	judgeCategory  []Item
	allJudges      []Item
	assignments    []Item
	allAssignments []Item
	reports        []Report
	report         Report
}

func (r *repositoryStub) RoleCodes(context.Context, int64) ([]string, error) { return r.roles, nil }
func (r *repositoryStub) Registration(context.Context) ([]Item, error)       { return r.registration, nil }
func (r *repositoryStub) AthleteRanks(context.Context, int64) ([]Item, error) {
	return r.athleteRanks, nil
}
func (r *repositoryStub) CoachAthleteRanks(context.Context, int64) ([]Item, error) {
	return r.coachRanks, nil
}
func (r *repositoryStub) AllRanks(context.Context) ([]Item, error) { return r.allRanks, nil }
func (r *repositoryStub) JudgeCategories(context.Context, int64) ([]Item, error) {
	return r.judgeCategory, nil
}
func (r *repositoryStub) AllJudgeCategories(context.Context) ([]Item, error) { return r.allJudges, nil }
func (r *repositoryStub) JudgeAssignments(context.Context, int64) ([]Item, error) {
	return r.assignments, nil
}
func (r *repositoryStub) AllJudgeAssignments(context.Context) ([]Item, error) {
	return r.allAssignments, nil
}
func (r *repositoryStub) ListReports(context.Context) ([]Report, error)    { return r.reports, nil }
func (r *repositoryStub) GetReport(context.Context, int64) (Report, error) { return r.report, nil }
func (r *repositoryStub) CreateReport(_ context.Context, actorID int64, name string, dueAt time.Time, status string) (Report, error) {
	return Report{ID: 1, Name: name, DueAt: dueAt, Status: status, CreatedByUserID: actorID}, nil
}
func (r *repositoryStub) UpdateReport(_ context.Context, report Report) (Report, error) {
	return report, nil
}
func (r *repositoryStub) CreateReportDeadlineReminders(context.Context, int32) error { return nil }

type permissionStub struct{ allowed bool }

func (p permissionStub) HasPermission(context.Context, int64, string) (bool, error) {
	return p.allowed, nil
}

func TestListParticipantDeadlines(t *testing.T) {
	due := time.Now().Add(time.Hour)
	repository := &repositoryStub{
		roles:        []string{"participant"},
		registration: []Item{{Type: "competition_registration", ID: 1, DueAt: due}},
		athleteRanks: []Item{{Type: "rank_expiration", ID: 2, DueAt: due.Add(time.Hour)}},
	}
	items, err := NewService(repository, permissionStub{}).List(context.Background(), 7)
	if err != nil || len(items) != 2 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}

func TestListFederationIncludesReports(t *testing.T) {
	repository := &repositoryStub{reports: []Report{{ID: 4, Name: "Отчёт", DueAt: time.Now()}}}
	items, err := NewService(repository, permissionStub{allowed: true}).List(context.Background(), 7)
	if err != nil || len(items) != 1 || items[0].Type != "report" {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}

func TestCreateReportDefaultsStatus(t *testing.T) {
	report, err := NewService(&repositoryStub{}, permissionStub{}).CreateReport(context.Background(), 7, CreateReportRequest{Name: " Ходатайства на премии ", DueAt: "2027-02-01"})
	if err != nil || report.Status != "planned" || report.Name != "Ходатайства на премии" {
		t.Fatalf("report=%#v err=%v", report, err)
	}
}

func TestListEmptyIsArray(t *testing.T) {
	items, err := NewService(&repositoryStub{}, permissionStub{}).List(context.Background(), 7)
	if err != nil || items == nil {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}
