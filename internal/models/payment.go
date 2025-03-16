package models

import (
    "github.com/google/uuid"
)

type Payment struct {
    Base
    OrderID      uuid.UUID `json:"order_id"`
    Order        Order     `gorm:"foreignKey:OrderID" json:"order"`
    Amount       float64   `gorm:"not null" json:"amount"`
    Currency     string    `gorm:"not null;default:'USD'" json:"currency"`
    Status       string    `gorm:"not null" json:"status"`
    Provider     string    `gorm:"not null" json:"provider"`
    PaymentToken string    `json:"payment_token"`
    ErrorMessage string    `json:"error_message,omitempty"`
}

type PaymentMethod struct {
    Base
    UserID      uuid.UUID `json:"user_id"`
    User        User      `gorm:"foreignKey:UserID" json:"user"`
    Type        string    `gorm:"not null" json:"type"` // credit_card, paypal, etc.
    Provider    string    `gorm:"not null" json:"provider"`
    TokenID     string    `gorm:"not null" json:"token_id"`
    Last4       string    `json:"last4,omitempty"`
    ExpiryMonth string    `json:"expiry_month,omitempty"`
    ExpiryYear  string    `json:"expiry_year,omitempty"`
    IsDefault   bool      `gorm:"default:false" json:"is_default"`
}