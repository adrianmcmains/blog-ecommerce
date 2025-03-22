// api/models/product.go
package models

import (
	"time"
)

// Product represents a product in the shop
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	SalePrice   float64   `json:"salePrice,omitempty"`
	SKU         string    `json:"sku"`
	Stock       int       `json:"stock"`
	Featured    bool      `json:"featured"`
	Visible     bool      `json:"visible"`
	Categories  []string  `json:"categories,omitempty"`
	Images      []string  `json:"images,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ProductCategory represents a product category
type ProductCategory struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Image       string    `json:"image,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}