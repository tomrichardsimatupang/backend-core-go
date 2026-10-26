package app

import (
	"testing"

	"backend-core-go/internal/features/rbac"
)

func TestPermissions_AreValid(t *testing.T) {
	perms, err := Permissions()
	if err != nil {
		t.Fatalf("Permissions: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("expected at least one permission")
	}
}

func TestDefaultRoles_UseKnownPermissions(t *testing.T) {
	perms, err := Permissions()
	if err != nil {
		t.Fatalf("Permissions: %v", err)
	}
	known := rbac.NewPermissionSet(perms...)
	for _, r := range DefaultRoles() {
		if r.Name == rbac.SuperAdminRole {
			t.Fatalf("%s is seeded automatically and must not be in DefaultRoles", rbac.SuperAdminRole)
		}
		for _, p := range r.Permissions {
			if !known.Has(p) {
				t.Errorf("role %q uses unknown permission %q", r.Name, p)
			}
		}
	}
}
