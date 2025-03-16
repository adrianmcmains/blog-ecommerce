package models

import (
    "github.com/google/uuid"
)

type Order struct {
    Base
    UserID      uuid.UUID    `json:"user_id"`
    User        User         `gorm:"foreignKey:UserID" json:"user"`
    Items       []OrderItem  `json:"items"`
    Status      string       `gorm:"not null;default:'pending'" json:"status"`
    TotalAmount float64      `gorm:"not null" json:"total_amount"`
    Address     Address      `gorm:"embedded" json:"address"`
    PaymentID   string       `json:"payment_id"`
}

type OrderItem struct {
    Base
    OrderID     uuid.UUID    `json:"order_id"`
    ProductID   uuid.UUID    `json:"product_id"`
    Product     Product      `gorm:"foreignKey:ProductID" json:"product"`
    VariantID   *uuid.UUID   `json:"variant_id,omitempty"`
    Variant     *ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
    Quantity    int          `gorm:"not null" json:"quantity"`
    PriceAtTime float64      `gorm:"not null" json:"price_at_time"`
}

type Address struct {
    Street     string `json:"street"`
    City       string `json:"city"`
    State      string `json:"state"`
    Country    string `json:"country"`
    PostalCode string `json:"postal_code"`
}