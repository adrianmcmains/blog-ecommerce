package models

import (
	"github.com/google/uuid"
	"time"
)

// Category represents a product category
type Category struct {
	ID        uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Slug      string    `gorm:"unique;not null" json:"slug"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null" json:"updated_at"`
	Products  []Product `gorm:"many2many:product_categories;" json:"products,omitempty"`
}