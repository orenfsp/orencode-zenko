package competition

import (
	"context"
	"errors"
	"testing"
	"time"
)

type judgeRepositoryStub struct {
	status        string
	statusErr     error
	items         []CompetitionJudge
	listErr       error
	validation    JudgeAssignmentValidation
	validationErr error
	created       CreatedCompetitionJudge
	assignErr     error
	removed       bool
	removeErr     error
	competitionID int64
	judgeUserID   int64
	positionID    int64
	assignmentID  int64
}

func (r *judgeRepositoryStub) GetCompetitionStatus(_ context.Context, competitionID int64) (string, error) {
	r.competitionID = competitionID
	return r.status, r.statusErr
}
func (r *judgeRepositoryStub) ListCompetitionJudges(_ context.Context, competitionID int64) ([]CompetitionJudge, error) {
	r.competitionID = competitionID
	return r.items, r.listErr
}
func (r *judgeRepositoryStub) ValidateCompetitionJudgeAssignment(
	_ context.Context, competitionID, judgeUserID, positionID int64,
) (JudgeAssignmentValidation, error) {
	r.competitionID, r.judgeUserID, r.positionID = competitionID, judgeUserID, positionID
	return r.validation, r.validationErr
}
func (r *judgeRepositoryStub) AssignCompetitionJudge(
	_ context.Context, competitionID, judgeUserID, positionID int64,
) (CreatedCompetitionJudge, error) {
	r.competitionID, r.judgeUserID, r.positionID = competitionID, judgeUserID, positionID
	return r.created, r.assignErr
}
func (r *judgeRepositoryStub) RemoveCompetitionJudge(
	_ context.Context, competitionID, assignmentID int64,
) (bool, error) {
	r.competitionID, r.assignmentID = competitionID, assignmentID
	return r.removed, r.removeErr
}

func validJudgeValidation() JudgeAssignmentValidation {
	return JudgeAssignmentValidation{
		CompetitionExists: true, UserExists: true, JudgeRoleExists: true, PositionExists: true,
	}
}

func TestJudgeServiceListsPublicCompetitionJudges(t *testing.T) {
	middleName := "Иванович"
	appointedAt := time.Date(2026, time.October, 1, 10, 0, 0, 0, time.UTC)
	repository := &judgeRepositoryStub{
		status: "published",
		items: []CompetitionJudge{{
			ID: 9, CompetitionID: 3, JudgeUserID: 42, FirstName: "Иван", LastName: "Иванов",
			MiddleName: &middleName, PositionID: 1, PositionName: "Главный судья", AppointedAt: appointedAt,
		}},
	}
	items, err := NewJudgeService(repository).ListPublic(context.Background(), 3)
	if err != nil {
		t.Fatalf("ListPublic() error = %v", err)
	}
	if len(items) != 1 || items[0].Judge.MiddleName == nil || *items[0].Judge.MiddleName != middleName ||
		items[0].Position.Name != "Главный судья" || !items[0].AppointedAt.Equal(appointedAt) {
		t.Fatalf("items = %#v", items)
	}
}

func TestJudgeServiceReturnsEmptyJudgeList(t *testing.T) {
	items, err := NewJudgeService(&judgeRepositoryStub{status: "completed"}).ListPublic(context.Background(), 3)
	if err != nil || items == nil || len(items) != 0 {
		t.Fatalf("ListPublic() = %#v, %v", items, err)
	}
}

func TestJudgeServiceHidesPrivateCompetition(t *testing.T) {
	_, err := NewJudgeService(&judgeRepositoryStub{status: "draft"}).ListPublic(context.Background(), 3)
	if !errors.Is(err, ErrJudgeCompetitionPrivate) {
		t.Fatalf("ListPublic() error = %v", err)
	}
}

func TestJudgeServiceAssignsJudge(t *testing.T) {
	appointedAt := time.Now()
	repository := &judgeRepositoryStub{
		status: "draft", validation: validJudgeValidation(),
		created: CreatedCompetitionJudge{
			ID: 8, CompetitionID: 3, JudgeUserID: 42, PositionID: 1, AppointedAt: appointedAt,
		},
	}
	created, err := NewJudgeService(repository).Assign(context.Background(), 3, AssignJudgeRequest{
		JudgeUserID: 42, PositionID: 1,
	})
	if err != nil {
		t.Fatalf("Assign() error = %v", err)
	}
	if created.ID != 8 || created.JudgeUserID != 42 || repository.positionID != 1 {
		t.Fatalf("created = %#v, repository = %#v", created, repository)
	}
}

func TestJudgeServiceRejectsInvalidAssignments(t *testing.T) {
	tests := []struct {
		name       string
		repository *judgeRepositoryStub
		want       error
	}{
		{name: "competition missing", repository: &judgeRepositoryStub{statusErr: ErrJudgeCompetitionMissing}, want: ErrJudgeCompetitionMissing},
		{name: "not draft", repository: &judgeRepositoryStub{status: "published"}, want: ErrJudgeCompetitionLocked},
		{name: "user missing", repository: &judgeRepositoryStub{status: "draft", validation: JudgeAssignmentValidation{CompetitionExists: true, PositionExists: true}}, want: ErrJudgeUserMissing},
		{name: "user without judge role", repository: &judgeRepositoryStub{status: "draft", validation: JudgeAssignmentValidation{CompetitionExists: true, UserExists: true, PositionExists: true}}, want: ErrUserIsNotJudge},
		{name: "position missing", repository: &judgeRepositoryStub{status: "draft", validation: JudgeAssignmentValidation{CompetitionExists: true, UserExists: true, JudgeRoleExists: true}}, want: ErrJudgePositionMissing},
		{name: "duplicate", repository: &judgeRepositoryStub{status: "draft", validation: JudgeAssignmentValidation{CompetitionExists: true, UserExists: true, JudgeRoleExists: true, PositionExists: true, AssignmentExists: true}}, want: ErrJudgeAssignmentExists},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewJudgeService(test.repository).Assign(context.Background(), 3, AssignJudgeRequest{JudgeUserID: 42, PositionID: 1})
			if !errors.Is(err, test.want) {
				t.Fatalf("Assign() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestJudgeServiceRemovesOwnedAssignment(t *testing.T) {
	repository := &judgeRepositoryStub{status: "draft", removed: true}
	if err := NewJudgeService(repository).Remove(context.Background(), 3, 8); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if repository.competitionID != 3 || repository.assignmentID != 8 {
		t.Fatalf("remove args = %d, %d", repository.competitionID, repository.assignmentID)
	}
}

func TestJudgeServiceRejectsForeignAssignment(t *testing.T) {
	err := NewJudgeService(&judgeRepositoryStub{status: "draft", removed: false}).Remove(context.Background(), 3, 8)
	if !errors.Is(err, ErrJudgeAssignmentMissing) {
		t.Fatalf("Remove() error = %v", err)
	}
}

func TestJudgeServicePropagatesRepositoryErrors(t *testing.T) {
	wantErr := errors.New("database unavailable")
	if _, err := NewJudgeService(&judgeRepositoryStub{status: "published", listErr: wantErr}).ListPublic(context.Background(), 3); !errors.Is(err, wantErr) {
		t.Fatalf("ListPublic() error = %v", err)
	}
	if _, err := NewJudgeService(&judgeRepositoryStub{status: "draft", validationErr: wantErr}).Assign(
		context.Background(), 3, AssignJudgeRequest{JudgeUserID: 42, PositionID: 1},
	); !errors.Is(err, wantErr) {
		t.Fatalf("Assign() error = %v", err)
	}
	if err := NewJudgeService(&judgeRepositoryStub{status: "draft", removeErr: wantErr}).Remove(context.Background(), 3, 8); !errors.Is(err, wantErr) {
		t.Fatalf("Remove() error = %v", err)
	}
}
