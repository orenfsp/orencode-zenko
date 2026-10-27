package result

import (
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

func TestPermissionsAreRegistered(t *testing.T) {
	registered := make(map[string]bool, len(rbac.DefaultPermissions))
	for _, permission := range rbac.DefaultPermissions {
		registered[permission.Code] = true
	}
	for _, permission := range []string{PermissionRead, PermissionManage, PermissionApprove} {
		if !registered[permission] {
			t.Fatalf("permission %q is not registered", permission)
		}
	}
}
