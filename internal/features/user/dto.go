package user

import "time"

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name" binding:"omitempty,min=2,max=100"`
	Email *string `json:"email" binding:"omitempty,email,max=255"`
}

type UserResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toResponse(u *User) UserResponse {
	return UserResponse{ID: u.ID, Name: u.Name, Email: u.Email, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
}

func toResponses(users []User) []UserResponse {
	out := make([]UserResponse, 0, len(users))
	for i := range users {
		out = append(out, toResponse(&users[i]))
	}
	return out
}
