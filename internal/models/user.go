package models

import (
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    Base
    Email        string    `gorm:"unique;not null" json:"email"`
    PasswordHash string    `gorm:"not null" json:"-"`
    Role         string    `gorm:"not null" json:"role"`
    FirstName    string    `json:"first_name"`
    LastName     string    `json:"last_name"`
    Posts        []Post    `json:"posts,omitempty"`
    Orders       []Order   `json:"orders,omitempty"`
    CartItems    []CartItem `json:"cart_items,omitempty"`
}

func (u *User) SetPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.PasswordHash = string(hashedPassword)
    return nil
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
    return err == nil
}

// ChangePasswordInput represents the input for changing a user's password
type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// SanitizeUser removes sensitive information from user for safe return
func SanitizeUser(user *User) map[string]interface{} {
	return map[string]interface{}{
		"id":         user.ID,
		"name":       user.FirstName,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}