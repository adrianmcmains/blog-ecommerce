package models

import (
    "github.com/google/uuid"
)

type Product struct {
    Base
    Name          string         `gorm:"not null" json:"name"`
    Slug          string         `gorm:"unique;not null" json:"slug"`
    Description   string         `gorm:"type:text" json:"description"`
    Price         float64        `gorm:"not null" json:"price"`
    StockQuantity int           `gorm:"not null" json:"stock_quantity"`
    Image         string         `json:"image"`
    Status        string         `gorm:"not null;default:'active'" json:"status"`
    Categories    []Category     `gorm:"many2many:product_categories;" json:"categories"`
    Variants      []ProductVariant `json:"variants,omitempty"`
    CartItems     []CartItem     `json:"cart_items,omitempty"`
}

type ProductVariant struct {
    Base
    ProductID     uuid.UUID `json:"product_id"`
    Name          string    `gorm:"not null" json:"name"`
    Price         float64   `gorm:"not null" json:"price"`
    StockQuantity int      `gorm:"not null" json:"stock_quantity"`
    SKU           string    `gorm:"unique" json:"sku"`
}

type CartItem struct {
    Base
    UserID       uuid.UUID `json:"user_id"`
    User         User      `gorm:"foreignKey:UserID" json:"-"`
    ProductID    uuid.UUID `json:"product_id"`
    Product      Product   `gorm:"foreignKey:ProductID" json:"product"`
    VariantID    *uuid.UUID `json:"variant_id,omitempty"`
    Variant      *ProductVariant `gorm:"foreignKey:VariantID" json:"variant,omitempty"`
    Quantity     int       `gorm:"not null" json:"quantity"`
}