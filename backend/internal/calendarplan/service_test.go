package calendarplan

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

type repositoryStub struct {
	plans             []CalendarPlan
	competitions      []Competition
	getPlan           CalendarPlan
	err               error
	listFilter        ListFilter
	competitionPlanID int64
	competitionLimit  int32
	competitionOffset int32
	created           CalendarPlan
	updated           UpdateCalendarPlan
}

func (r *repositoryStub) List(_ context.Context, filter ListFilter) ([]CalendarPlan, error) {
	r.listFilter = filter
	return r.plans, r.err
}
func (r *repositoryStub) Get(_ context.Context, _ int64) (CalendarPlan, error) {
	if r.err != nil {
		return CalendarPlan{}, r.err
	}
	return r.getPlan, nil
}
func (r *repositoryStub) Create(_ context.Context, plan CalendarPlan) (CalendarPlan, error) {
	r.created = plan
	if r.err != nil {
		return CalendarPlan{}, r.err
	}
	plan.ID = 9
	return plan, nil
}
func (r *repositoryStub) Update(_ context.Context, plan UpdateCalendarPlan) (CalendarPlan, error) {
	r.updated = plan
	if r.err != nil {
		return CalendarPlan{}, r.err
	}
	return CalendarPlan{ID: plan.ID, Name: plan.Name, Year: plan.Year}, nil
}
func (r *repositoryStub) ListCompetitions(_ context.Context, planID int64, limit, offset int32) ([]Competition, error) {
	r.competitionPlanID, r.competitionLimit, r.competitionOffset = planID, limit, offset
	return r.competitions, r.err
}

func TestServiceListPlansWithYearAndPagination(t *testing.T) {
	year := int16(2026)
	repo := &repositoryStub{plans: []CalendarPlan{{ID: 1, Name: "Областной план", Year: year}}}
	items, err := NewService(repo).List(context.Background(), ListFilter{Year: &year, Limit: 20, Offset: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 1 || items[0].Name != "Областной план" {
		t.Fatalf("List() = %#v", items)
	}
	if repo.listFilter.Year == nil || *repo.listFilter.Year != year || repo.listFilter.Limit != 20 || repo.listFilter.Offset != 10 {
		t.Fatalf("filter = %#v", repo.listFilter)
	}
}

func TestServiceListReturnsEmptySlice(t *testing.T) {
	items, err := NewService(&repositoryStub{}).List(context.Background(), ListFilter{Limit: DefaultLimit})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("List() = %#v, want non-nil empty slice", items)
	}
}

func TestServiceListRejectsInvalidFilters(t *testing.T) {
	invalidYear := int16(0)
	for _, filter := range []ListFilter{
		{Limit: 0}, {Limit: MaxLimit + 1}, {Limit: 1, Offset: -1}, {Year: &invalidYear, Limit: 1},
	} {
		if _, err := NewService(&repositoryStub{}).List(context.Background(), filter); err == nil {
			t.Fatalf("List(%#v) expected error", filter)
		}
	}
}

func TestServiceGetsPlanAndMissingID(t *testing.T) {
	repo := &repositoryStub{getPlan: CalendarPlan{ID: 7, Name: "ЕКП", Year: 2026}}
	item, err := NewService(repo).Get(context.Background(), 7)
	if err != nil || item.ID != 7 {
		t.Fatalf("Get() = %#v, %v", item, err)
	}
	repo.err = ErrNotFound
	if _, err := NewService(repo).Get(context.Background(), 8); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get() error = %v", err)
	}
}

func TestServiceCreatesNormalizedPlan(t *testing.T) {
	repo := &repositoryStub{}
	item, err := NewService(repo).Create(context.Background(), CreateRequest{Name: "  План области  ", Year: 2026})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.ID != 9 || repo.created.Name != "План области" || repo.created.Year != 2026 {
		t.Fatalf("created = %#v, response = %#v", repo.created, item)
	}
}

func TestServiceCreateValidation(t *testing.T) {
	for _, request := range []CreateRequest{
		{Name: "   ", Year: 2026},
		{Name: strings.Repeat("я", MaxNameRunes+1), Year: 2026},
		{Name: "План", Year: 0},
		{Name: "План", Year: -1},
	} {
		if _, err := NewService(&repositoryStub{}).Create(context.Background(), request); !errors.Is(err, ErrInvalidCreate) {
			t.Fatalf("Create(%#v) error = %v", request, err)
		}
	}
}

func TestServiceUpdatesOnlyProvidedField(t *testing.T) {
	repo := &repositoryStub{getPlan: CalendarPlan{ID: 4, Name: "Старое имя", Year: 2025}}
	request := UpdateRequest{Name: PatchField[string]{Present: true, Value: "  Новое имя "}}
	item, err := NewService(repo).Update(context.Background(), 4, request)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	want := UpdateCalendarPlan{ID: 4, Name: "Новое имя", Year: 2025}
	if !reflect.DeepEqual(repo.updated, want) || item.Name != want.Name || item.Year != want.Year {
		t.Fatalf("updated = %#v, response = %#v", repo.updated, item)
	}
}

func TestServiceRejectsEmptyNullAndInvalidPatch(t *testing.T) {
	repo := &repositoryStub{getPlan: CalendarPlan{ID: 4, Name: "План", Year: 2025}}
	requests := []UpdateRequest{
		{},
		{Name: PatchField[string]{Present: true, Null: true}},
		{Year: PatchField[int16]{Present: true, Null: true}},
		{Name: PatchField[string]{Present: true, Value: " "}},
		{Year: PatchField[int16]{Present: true, Value: 0}},
	}
	for _, request := range requests {
		if _, err := NewService(repo).Update(context.Background(), 4, request); !errors.Is(err, ErrInvalidUpdate) {
			t.Fatalf("Update(%#v) error = %v", request, err)
		}
	}
}

func TestServiceListsAllPlanCompetitionsIncludingDrafts(t *testing.T) {
	number := "12"
	now := time.Now()
	repo := &repositoryStub{
		getPlan: CalendarPlan{ID: 2, Name: "План", Year: 2026},
		competitions: []Competition{
			{ID: 1, Name: "Черновик", Status: "draft", StartAt: now, EndAt: now.Add(time.Hour), DisciplineID: 3, DisciplineName: "Алгоритмическое программирование", CalendarPlanNumber: &number},
		},
	}
	items, err := NewService(repo).ListCompetitions(context.Background(), 2, 25, 5)
	if err != nil {
		t.Fatalf("ListCompetitions() error = %v", err)
	}
	if len(items) != 1 || items[0].Status != "draft" || items[0].Discipline.ID != 3 || items[0].CalendarPlanNumber == nil {
		t.Fatalf("items = %#v", items)
	}
	if repo.competitionPlanID != 2 || repo.competitionLimit != 25 || repo.competitionOffset != 5 {
		t.Fatalf("repository args = %d, %d, %d", repo.competitionPlanID, repo.competitionLimit, repo.competitionOffset)
	}
}

func TestServicePlanWithoutCompetitionsReturnsEmptySlice(t *testing.T) {
	repo := &repositoryStub{getPlan: CalendarPlan{ID: 2}}
	items, err := NewService(repo).ListCompetitions(context.Background(), 2, DefaultLimit, 0)
	if err != nil || items == nil || len(items) != 0 {
		t.Fatalf("ListCompetitions() = %#v, %v", items, err)
	}
}

func TestServicePropagatesRepositoryErrors(t *testing.T) {
	repositoryError := errors.New("database unavailable")
	repo := &repositoryStub{err: repositoryError}
	service := NewService(repo)
	if _, err := service.List(context.Background(), ListFilter{Limit: 1}); !errors.Is(err, repositoryError) {
		t.Fatalf("List() error = %v", err)
	}
	if _, err := service.Create(context.Background(), CreateRequest{Name: "План", Year: 2026}); !errors.Is(err, repositoryError) {
		t.Fatalf("Create() error = %v", err)
	}
	if _, err := service.ListCompetitions(context.Background(), 1, 1, 0); !errors.Is(err, repositoryError) {
		t.Fatalf("ListCompetitions() error = %v", err)
	}
}
