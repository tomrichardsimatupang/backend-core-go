package rbac

import "time"

type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=50"`
	Description string `json:"description" binding:"max=255"`
}

type SetPermissionsRequest struct {
	Permissions []string `json:"permissions" binding:"required,dive,required"`
}

type AssignRoleRequest struct {
	RoleID int64 `json:"role_id" binding:"required,gt=0"`
}

type RoleResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type PermissionResponse struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

func toRoleResponse(r *Role) RoleResponse {
	return RoleResponse{ID: r.ID, Name: r.Name, Description: r.Description, CreatedAt: r.CreatedAt}
}

func toRoleResponses(roles []Role) []RoleResponse {
	out := make([]RoleResponse, 0, len(roles))
	for i := range roles {
		out = append(out, toRoleResponse(&roles[i]))
	}
	return out
}

func toPermissionResponses(perms []Permission) []PermissionResponse {
	out := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		out = append(out, PermissionResponse{Code: p.Code, Description: p.Description})
	}
	return out
}
