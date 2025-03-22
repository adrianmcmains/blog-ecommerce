package auth

import (
	//"time"

	//"github.com/google/uuid"
)

// UserResponse is what's returned to clients
/**type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}**/

// ToResponse converts User to UserResponse (hiding sensitive fields)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// LoginRequest represents the login form data
/**type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}**/

// RegisterRequest represents the registration form data
/**type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}**/
