package main

import (
	"reflect"
	"testing"

	"github.com/orenfsp/orencode-zenko.git/internal/integration"
)

func TestParsePermissions(t *testing.T) {
	permissions, err := parsePermissions(integration.PermissionWebhooksManage + "," + integration.PermissionResultsImport)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{integration.PermissionWebhooksManage, integration.PermissionResultsImport}
	if !reflect.DeepEqual(permissions, want) {
		t.Fatalf("permissions = %v", permissions)
	}
}

func TestParsePermissionsRejectsUnknown(t *testing.T) {
	if _, err := parsePermissions("unknown"); err == nil {
		t.Fatal("expected error")
	}
}
