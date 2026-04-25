package auth

import (
	"testing"
)

func TestHasPermission(t *testing.T) {
	if !HasPermission(RoleSuperAdmin, PermWriteDevices) {
		t.Error("expected super_admin to have devices:write")
	}
	if HasPermission(RoleDriver, PermWriteDevices) {
		t.Error("expected driver NOT to have devices:write")
	}
}
