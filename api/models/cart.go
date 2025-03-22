// api/models/cart.go
package models

import (
	"time"
)

// Cart represents a user's shopping cart
type Cart struct {
	ID        int         `json:"id"`
	UserID    int         `json:"userId"`
	Items     []CartItem  `json:"items"`
	Total     float64     `json:"total"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}

// CartItem represents an item in a shopping cart
type CartItem struct {
	ID        int       `json:"id"`
	CartID    int       `json:"cartId"`
	ProductID int       `json:"productId"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	Quantity  int       `json:"quantity"`
	Total     float64   `json:"total"`
	Image     string    `json:"image,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}