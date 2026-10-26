package rbac

import (
	"slices"
	"testing"
)

func TestAllPermissions_DedupesAndSorts(t *testing.T) {
	got, err := AllPermissions([]string{"user:read", "role:read"}, []string{"role:read", "product:manage"})
	if err != nil {
		t.Fatalf("AllPermissions: %v", err)
	}
	want := []string{"product:manage", "role:read", "user:read"}
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestAllPermissions_RejectsInvalidFormat(t *testing.T) {
	for _, code := range []string{"User:read", "user", "user:", ":read", "user:read:all", "user read"} {
		if _, err := AllPermissions([]string{code}); err == nil {
			t.Errorf("expected error for %q", code)
		}
	}
}
