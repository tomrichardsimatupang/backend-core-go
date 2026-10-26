package rbac

import "testing"

func TestValidateRoleSeeds(t *testing.T) {
	perms := []string{"post:read", "post:create"}
	tests := []struct {
		name    string
		roles   []RoleSeed
		wantErr bool
	}{
		{"valid", []RoleSeed{{Name: "author", Permissions: []string{"post:read", "post:create"}}}, false},
		{"unknown permission", []RoleSeed{{Name: "author", Permissions: []string{"post:fly"}}}, true},
		{"duplicate role", []RoleSeed{{Name: "author"}, {Name: "author"}}, true},
		{"reserved super_admin", []RoleSeed{{Name: SuperAdminRole}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoleSeeds(perms, tt.roles)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
