package competition

import (
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/rbac"
)

func TestJudgePermissionIsRegisteredForRBACSeed(t *testing.T) {
	if PermissionManageJudges != rbac.PermissionCompetitionJudgesManage {
		t.Fatalf("permission constants differ: %q and %q", PermissionManageJudges, rbac.PermissionCompetitionJudgesManage)
	}
	for _, permission := range rbac.DefaultPermissions {
		if permission.Code == PermissionManageJudges {
			return
		}
	}
	t.Fatalf("permission %q is not registered in rbac.DefaultPermissions", PermissionManageJudges)
}
