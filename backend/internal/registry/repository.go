package registry

import "context"

type Repository interface {
	ListUsers(ctx context.Context, filter UserFilter) (Page[User], error)
	ListOrganizations(ctx context.Context, filter OrganizationFilter) (Page[Organization], error)
}
