package models

import "time"

// User represents a user in the system
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"-"` // Password is never exposed in JSON responses
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// UserRole represents the possible roles for users
type UserRole string

const (
	// RoleAdmin represents an administrator user
	RoleAdmin UserRole = "admin"
	
	// RoleCustomer represents a regular customer
	RoleCustomer UserRole = "customer"
	
	// RoleContributor represents a blog contributor
	RoleContributor UserRole = "contributor"
)