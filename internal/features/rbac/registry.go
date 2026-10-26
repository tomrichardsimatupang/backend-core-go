package rbac

import (
	"fmt"
	"regexp"
	"sort"
)

var permissionPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*$`)

// AllPermissions menggabungkan slice Permissions dari semua feature menjadi
// daftar unik yang terurut, dan memvalidasi format resource:action.
//
// rbac tidak bisa mengimpor feature lain (feature lain mengimpor rbac untuk
// middleware), jadi daftar feature dirangkai di internal/app/permissions.go.
func AllPermissions(groups ...[]string) ([]string, error) {
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, group := range groups {
		for _, code := range group {
			if !permissionPattern.MatchString(code) {
				return nil, fmt.Errorf("invalid permission code %q: must be lowercase resource:action", code)
			}
			if _, ok := seen[code]; ok {
				continue
			}
			seen[code] = struct{}{}
			out = append(out, code)
		}
	}
	sort.Strings(out)
	return out, nil
}
