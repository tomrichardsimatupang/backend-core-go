package app

import "testing"

func TestPermissions_AreValid(t *testing.T) {
	perms, err := Permissions()
	if err != nil {
		t.Fatalf("Permissions: %v", err)
	}
	if len(perms) == 0 {
		t.Fatal("expected at least one permission")
	}
}
