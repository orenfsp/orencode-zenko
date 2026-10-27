package antidoping

import (
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

func TestReviewPermissionIsRegisteredForRBACSeed(t *testing.T) {
	for _, permission := range rbac.DefaultPermissions {
		if permission.Code == rbac.PermissionAntiDopingReview {
			return
		}
	}
	t.Fatalf("permission %q is not registered", rbac.PermissionAntiDopingReview)
}
