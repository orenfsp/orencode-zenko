package rbac

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct {
	allowed       bool
	permissionErr error
	createRequest CreateRoleRequest
}

func (r *repositoryStub) HasPermission(context.Context, int64, string) (bool, error) {
	return r.allowed, r.permissionErr
}

func (r *repositoryStub) CreateRole(_ context.Context, request CreateRoleRequest) (Role, error) {
	r.createRequest = request
	return Role{ID: 1, Code: request.Code, Name: request.Name, Permissions: request.PermissionCodes}, nil
}

func (*repositoryStub) AssignRole(context.Context, int64, int64) error { return nil }
func (*repositoryStub) ListRoles(context.Context) ([]Role, error)      { return nil, nil }
func (*repositoryStub) ListPermissions(context.Context) ([]Permission, error) {
	return nil, nil
}
func (*repositoryStub) ListSelfAssignableRoles(context.Context) ([]Role, error) {
	return nil, nil
}
func (*repositoryStub) ListRolePermissions(context.Context, int64) ([]string, error) {
	return nil, nil
}

func TestCreateRoleNormalizesAndDeduplicatesPermissions(t *testing.T) {
	repository := &repositoryStub{}
	service := NewService(repository)

	role, err := service.CreateRole(context.Background(), CreateRoleRequest{
		Code:            " Coach.Manager ",
		Name:            " Тренер-менеджер ",
		PermissionCodes: []string{"roles.read", " roles.assign ", "roles.read"},
	})
	if err != nil {
		t.Fatalf("CreateRole() error = %v", err)
	}
	if role.Code != "coach.manager" {
		t.Fatalf("role code = %q", role.Code)
	}
	if len(repository.createRequest.PermissionCodes) != 2 {
		t.Fatalf("permissions = %#v", repository.createRequest.PermissionCodes)
	}
}

func TestCreateRoleRejectsInvalidCode(t *testing.T) {
	service := NewService(&repositoryStub{})
	_, err := service.CreateRole(context.Background(), CreateRoleRequest{Code: "ADMIN ROLE", Name: "Admin"})
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("CreateRole() error = %v, want ErrInvalidRole", err)
	}
}
