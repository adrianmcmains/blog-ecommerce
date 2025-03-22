// File: api/auth/models.go
package auth

import (
	"time"

	"github.com/google/uuid"
	//"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID        uuid.UUID `json:"id" gorm:"primary_key;type:uuid;default:gen_random_uuid()"`
	Email     string    `json:"email" gorm:"unique;not null"`
	Password  string    `json:"-" gorm:"not null"` // Password never exposed in JSON
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role" gorm:"default:'customer'"` // admin, customer, contributor
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

// UserResponse is what's returned to clients
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// ToResponse converts User to UserResponse (hiding sensitive fields)
/**func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}**/

// LoginRequest represents the login form data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest represents the registration form data
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

// CheckPasswordHash compares password with hash
/**func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}**/

// HashPassword hashes a password
/**func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}**/
