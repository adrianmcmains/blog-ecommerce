// File: api/models/order.go

package models

import (
	"time"
)

// Order statuses
const (
	OrderStatusPending        = "pending"
	OrderStatusPaymentPending = "payment_pending"
	OrderStatusPaid           = "paid"
	OrderStatusProcessing     = "processing"
	OrderStatusShipped        = "shipped"
	OrderStatusDelivered      = "delivered"
	OrderStatusCanceled       = "canceled"
	OrderStatusRefunded       = "refunded"
	OrderStatusPaymentFailed  = "payment_failed"
	OrderStatusPaymentCanceled = "payment_canceled"
)

// Order represents an order in the system
type Order struct {
	ID              int         `json:"id"`
	UserID          int         `json:"userId"`
	TotalAmount     float64     `json:"totalAmount"`
	Status          string      `json:"status"`
	TransactionID   string      `json:"transactionId,omitempty"`
	ShippingAddress string      `json:"shippingAddress"`
	BillingAddress  string      `json:"billingAddress"`
	PaymentMethod   string      `json:"paymentMethod"`
	Notes           string      `json:"notes,omitempty"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
	Items           []OrderItem `json:"items,omitempty"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID          int       `json:"id"`
	OrderID     int       `json:"orderId"`
	ProductID   int       `json:"productId"`
	ProductName string    `json:"productName"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unitPrice"`
	TotalPrice  float64   `json:"totalPrice"`
	CreatedAt   time.Time `json:"createdAt"`
}