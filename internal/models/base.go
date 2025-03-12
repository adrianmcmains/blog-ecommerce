package models

import (
    "time"

    "github.com/google/uuid"
    "gorm.io/gorm"
)

// Base contains common fields for all models
type Base struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (base *Base) BeforeCreate(tx *gorm.DB) error {
    if base.ID == uuid.Nil {
        base.ID = uuid.New()
    }
    return nil
}