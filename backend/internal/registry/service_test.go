package registry

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	users              Page[User]
	organizations      Page[Organization]
	err                error
	userFilter         UserFilter
	organizationFilter OrganizationFilter
}

func (r *repositoryStub) ListUsers(_ context.Context, filter UserFilter) (Page[User], error) {
	r.userFilter = filter
	return r.users, r.err
}

func (r *repositoryStub) ListOrganizations(_ context.Context, filter OrganizationFilter) (Page[Organization], error) {
	r.organizationFilter = filter
	return r.organizations, r.err
}

func TestListUsersNormalizesSearchAndPassesFilters(t *testing.T) {
	search, role := "  Иванов ", " participant "
	repository := &repositoryStub{users: Page[User]{Items: []User{{ID: 7}}}}
	page, err := NewService(repository).ListUsers(context.Background(), UserFilter{
		Search: &search, RoleCode: &role, Limit: 25,
	})
	if err != nil {
		t.Fatal(err)
	}
	if page.Items[0].ID != 7 || repository.userFilter.Search == nil || *repository.userFilter.Search != "Иванов" {
		t.Fatalf("page=%#v filter=%#v", page, repository.userFilter)
	}
	if repository.userFilter.RoleCode == nil || *repository.userFilter.RoleCode != "participant" {
		t.Fatalf("role = %#v", repository.userFilter.RoleCode)
	}
}

func TestListUsersRejectsInvalidPagination(t *testing.T) {
	_, err := NewService(&repositoryStub{}).ListUsers(context.Background(), UserFilter{Limit: 101})
	if !errors.Is(err, ErrInvalidFilter) {
		t.Fatalf("error=%v", err)
	}
}

func TestListOrganizationsReturnsRepositoryError(t *testing.T) {
	want := errors.New("database unavailable")
	_, err := NewService(&repositoryStub{err: want}).ListOrganizations(context.Background(), OrganizationFilter{Limit: 50})
	if !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}
}

func TestDecodeJSONValue(t *testing.T) {
	var roles []Role
	if err := decodeJSONValue([]any{map[string]any{"id": float64(1), "code": "judge", "name": "Судья"}}, &roles); err != nil {
		t.Fatal(err)
	}
	if len(roles) != 1 || roles[0].Code != "judge" {
		t.Fatalf("roles=%#v", roles)
	}
}
