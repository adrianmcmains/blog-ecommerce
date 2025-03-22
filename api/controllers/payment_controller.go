package controllers

import (
	"database/sql"
	"io"

	"github.com/adrianmcmains/blog-ecommerce/api/models"
	"github.com/adrianmcmains/blog-ecommerce/api/utils"

	//"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentController handles payment-related requests
type PaymentController struct {
	db             *sql.DB
	paymentService *utils.EversendPaymentService
	emailService   *utils.EmailService
}

// NewPaymentController creates a new payment controller
func NewPaymentController(db *sql.DB) (*PaymentController, error) {
	// Initialize Eversend payment service
	paymentService, err := utils.NewEversendPaymentService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize payment service: %w", err)
	}

	// Initialize email service
	emailService, err := utils.NewEmailService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize email service: %w", err)
	}

	return &PaymentController{
		db:             db,
		paymentService: paymentService,
		emailService:   emailService,
	}, nil
}

// CreatePaymentRequest represents a request to create a payment
type CreatePaymentRequest struct {
	OrderID     int               `json:"orderId" binding:"required"`
	Currency    string            `json:"currency" binding:"required"`
	RedirectURL string            `json:"redirectUrl" binding:"required"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// PaymentResponse represents a payment response
type PaymentResponse struct {
	Success      bool      `json:"success"`
	PaymentID    string    `json:"paymentId,omitempty"`
	PaymentURL   string    `json:"paymentUrl,omitempty"`
	Status       string    `json:"status,omitempty"`
	Reference    string    `json:"reference,omitempty"`
	Amount       float64   `json:"amount,omitempty"`
	Currency     string    `json:"currency,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
	UpdatedAt    time.Time `json:"updatedAt,omitempty"`
	ErrorMessage string    `json:"errorMessage,omitempty"`
}

// InitiatePayment handles the initiation of a payment for an order
func (c *PaymentController) InitiatePayment(ctx *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var req CreatePaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get order from database
	var order models.Order
	err := c.db.QueryRow(`
		SELECT id, user_id, total_amount, status
		FROM orders
		WHERE id = $1
	`, req.OrderID).Scan(&order.ID, &order.UserID, &order.TotalAmount, &order.Status)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify that the order belongs to the authenticated user
	if order.UserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this order"})
		return
	}

	// Check if order is in a valid state for payment
	if order.Status != "pending" && order.Status != "payment_failed" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Order is not in a valid state for payment"})
		return
	}

	// Get customer information
	var customer models.User
	err = c.db.QueryRow(`
		SELECT id, name, email
		FROM users
		WHERE id = $1
	`, userID).Scan(&customer.ID, &customer.Name, &customer.Email)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get customer information"})
		return
	}

	// Split name into first name and last name (assuming name is "First Last")
	firstName, lastName := splitName(customer.Name)

	// Generate a unique reference for the payment
	reference := fmt.Sprintf("order_%d_%s", order.ID, uuid.New().String()[:8])

	// Get base URL from environment or use default
	baseURL := os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080/api"
	}

	// Create webhook URL
	webhookURL := fmt.Sprintf("%s/payments/webhook", baseURL)

	// Prepare payment request
	paymentReq := utils.EversendPaymentRequest{
		Amount:      order.TotalAmount,
		Currency:    req.Currency,
		Description: fmt.Sprintf("Payment for Order #%d", order.ID),
		Reference:   reference,
		CustomerInfo: utils.EversendCustomer{
			FirstName: firstName,
			LastName:  lastName,
			Email:     customer.Email,
		},
		RedirectURL: req.RedirectURL,
		WebhookURL:  webhookURL,
		Metadata: map[string]string{
			"orderID": strconv.Itoa(order.ID),
			"userID":  strconv.Itoa(customer.ID),
		},
	}

	// If custom metadata is provided, merge it
	if req.Metadata != nil {
		for k, v := range req.Metadata {
			paymentReq.Metadata[k] = v
		}
	}

	// Create payment
	paymentResp, err := c.paymentService.CreatePayment(paymentReq)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create payment: %v", err)})
		return
	}

	// Update order status
	_, err = c.db.Exec(`
		UPDATE orders
		SET status = $1, transaction_id = $2, updated_at = NOW()
		WHERE id = $3
	`, "payment_pending", paymentResp.PaymentID, order.ID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// Store payment details in payment_transactions table
	_, err = c.db.Exec(`
		INSERT INTO payment_transactions (
			order_id, 
			transaction_id, 
			provider, 
			amount, 
			currency, 
			status, 
			payment_url,
			reference,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`, order.ID, paymentResp.PaymentID, "eversend", order.TotalAmount, req.Currency, paymentResp.Status, paymentResp.PaymentURL, reference)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store payment transaction"})
		return
	}

	// Return payment information
	ctx.JSON(http.StatusOK, PaymentResponse{
		Success:   true,
		PaymentID: paymentResp.PaymentID,
		PaymentURL: paymentResp.PaymentURL,
		Status:    paymentResp.Status,
		Reference: paymentResp.Reference,
		Amount:    order.TotalAmount,
		Currency:  req.Currency,
	})
}

// GetPaymentStatus gets the status of a payment
func (c *PaymentController) GetPaymentStatus(ctx *gin.Context) {
	// Get payment ID from URL parameter
	paymentID := ctx.Param("id")
	if paymentID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	// Check if the payment exists in our database
	var orderID int
	var userID int
	var status string
	err := c.db.QueryRow(`
		SELECT t.order_id, o.user_id, t.status
		FROM payment_transactions t
		JOIN orders o ON t.order_id = o.id
		WHERE t.transaction_id = $1
	`, paymentID).Scan(&orderID, &userID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if the authenticated user owns the payment
	authUserID, exists := ctx.Get("userId")
	if !exists || authUserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this payment"})
		return
	}

	// Get payment status from Eversend
	paymentStatus, err := c.paymentService.GetPaymentStatus(paymentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to get payment status: %v", err)})
		return
	}

	// Check if status has changed
	if paymentStatus.Status != status {
		// Update status in our database
		_, err = c.db.Exec(`
			UPDATE payment_transactions
			SET status = $1, updated_at = NOW()
			WHERE transaction_id = $2
		`, paymentStatus.Status, paymentID)
		
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
			return
		}

		// Update order status if payment is completed or failed
		if paymentStatus.Status == "completed" {
			_, err = c.db.Exec(`
				UPDATE orders
				SET status = 'paid', updated_at = NOW()
				WHERE id = $1
			`, orderID)
			
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
				return
			}

			// Send order confirmation email
			go c.sendOrderConfirmationEmail(orderID)
		} else if paymentStatus.Status == "failed" {
			_, err = c.db.Exec(`
				UPDATE orders
				SET status = 'payment_failed', updated_at = NOW()
				WHERE id = $1
			`, orderID)
			
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
				return
			}
		}
	}

	// Return payment status
	ctx.JSON(http.StatusOK, PaymentResponse{
		Success:   true,
		PaymentID: paymentStatus.PaymentID,
		Status:    paymentStatus.Status,
		Reference: paymentStatus.Reference,
		Amount:    paymentStatus.Amount,
		Currency:  paymentStatus.Currency,
		CreatedAt: paymentStatus.CreatedAt,
		UpdatedAt: paymentStatus.UpdatedAt,
	})
}

// WebhookHandler handles webhooks from Eversend
func (c *PaymentController) WebhookHandler(ctx *gin.Context) {
	// Get request body
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get signature from header
	signature := ctx.GetHeader("X-Eversend-Signature")
	if signature == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing webhook signature"})
		return
	}

	// Verify signature
	if !c.paymentService.VerifyWebhookSignature(body, signature) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Parse webhook payload
	webhookPayload, err := c.paymentService.ParseWebhookPayload(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to parse webhook payload: %v", err)})
		return
	}

	// Get order ID from metadata
	orderIDStr, ok := webhookPayload.Metadata["orderID"]
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing orderID in metadata"})
		return
	}

	orderID, err := strconv.Atoi(orderIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid orderID in metadata"})
		return
	}

	// Update payment transaction status
	_, err = c.db.Exec(`
		UPDATE payment_transactions
		SET status = $1, updated_at = NOW()
		WHERE transaction_id = $2
	`, webhookPayload.Status, webhookPayload.PaymentID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// Update order status based on payment status
	if webhookPayload.Status == "completed" {
		_, err = c.db.Exec(`
			UPDATE orders
			SET status = 'paid', updated_at = NOW()
			WHERE id = $1
		`, orderID)
		
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}

		// Send order confirmation email
		go c.sendOrderConfirmationEmail(orderID)
	} else if webhookPayload.Status == "failed" {
		_, err = c.db.Exec(`
			UPDATE orders
			SET status = 'payment_failed', updated_at = NOW()
			WHERE id = $1
		`, orderID)
		
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
			return
		}
	}

	// Return success response
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

// CancelPayment cancels a pending payment
func (c *PaymentController) CancelPayment(ctx *gin.Context) {
	// Get payment ID from URL parameter
	paymentID := ctx.Param("id")
	if paymentID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payment ID is required"})
		return
	}

	// Check if the payment exists in our database
	var orderID int
	var userID int
	var status string
	err := c.db.QueryRow(`
		SELECT t.order_id, o.user_id, t.status
		FROM payment_transactions t
		JOIN orders o ON t.order_id = o.id
		WHERE t.transaction_id = $1
	`, paymentID).Scan(&orderID, &userID, &status)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if the authenticated user owns the payment
	authUserID, exists := ctx.Get("userId")
	if !exists || authUserID != userID {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to access this payment"})
		return
	}

	// Check if payment is in a state that can be canceled
	if status != "pending" && status != "created" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payment cannot be canceled in its current state"})
		return
	}

	// Cancel payment in Eversend
	err = c.paymentService.CancelPayment(paymentID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to cancel payment: %v", err)})
		return
	}

	// Update payment transaction status
	_, err = c.db.Exec(`
		UPDATE payment_transactions
		SET status = 'canceled', updated_at = NOW()
		WHERE transaction_id = $1
	`, paymentID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	// Update order status
	_, err = c.db.Exec(`
		UPDATE orders
		SET status = 'payment_canceled', updated_at = NOW()
		WHERE id = $1
	`, orderID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	// Return success response
	ctx.JSON(http.StatusOK, gin.H{"success": true, "message": "Payment canceled successfully"})
}

// Helper function to split full name into first and last name
func splitName(fullName string) (string, string) {
	var firstName, lastName string
	
	// Find the first space in the name
	spaceIndex := -1
	for i, r := range fullName {
		if r == ' ' {
			spaceIndex = i
			break
		}
	}
	
	// If no space is found, use the full name as first name
	if spaceIndex == -1 {
		firstName = fullName
		lastName = ""
	} else {
		firstName = fullName[:spaceIndex]
		lastName = fullName[spaceIndex+1:]
	}
	
	return firstName, lastName
}

// sendOrderConfirmationEmail sends an order confirmation email
func (c *PaymentController) sendOrderConfirmationEmail(orderID int) {
	var (
		userID       int
		userEmail    string
		userName     string
		orderNumber  int
		totalAmount  float64
		orderDetails string
	)

	// Get order and user information
	err := c.db.QueryRow(`
		SELECT o.id, o.user_id, o.total_amount, u.email, u.name
		FROM orders o
		JOIN users u ON o.user_id = u.id
		WHERE o.id = $1
	`, orderID).Scan(&orderNumber, &userID, &totalAmount, &userEmail, &userName)

	if err != nil {
		fmt.Printf("Error getting order information for email: %v\n", err)
		return
	}

	// Get order items
	rows, err := c.db.Query(`
		SELECT product_name, quantity, unit_price, total_price
		FROM order_items
		WHERE order_id = $1
		ORDER BY id
	`, orderID)

	if err != nil {
		fmt.Printf("Error getting order items for email: %v\n", err)
		return
	}
	defer rows.Close()

	// Build order details HTML
	orderDetails = "<table style='width:100%; border-collapse:collapse;'>"
	orderDetails += "<tr style='background-color:#f0f0f0;'><th style='padding:8px; text-align:left; border:1px solid #ddd;'>Product</th><th style='padding:8px; text-align:right; border:1px solid #ddd;'>Quantity</th><th style='padding:8px; text-align:right; border:1px solid #ddd;'>Unit Price</th><th style='padding:8px; text-align:right; border:1px solid #ddd;'>Total</th></tr>"

	for rows.Next() {
		var productName string
		var quantity int
		var unitPrice, totalPrice float64

		if err := rows.Scan(&productName, &quantity, &unitPrice, &totalPrice); err != nil {
			fmt.Printf("Error scanning order item: %v\n", err)
			continue
		}

		orderDetails += fmt.Sprintf("<tr><td style='padding:8px; border:1px solid #ddd;'>%s</td><td style='padding:8px; text-align:right; border:1px solid #ddd;'>%d</td><td style='padding:8px; text-align:right; border:1px solid #ddd;'>$%.2f</td><td style='padding:8px; text-align:right; border:1px solid #ddd;'>$%.2f</td></tr>", productName, quantity, unitPrice, totalPrice)
	}

	orderDetails += fmt.Sprintf("<tr style='font-weight:bold;'><td colspan='3' style='padding:8px; text-align:right; border:1px solid #ddd;'>Total</td><td style='padding:8px; text-align:right; border:1px solid #ddd;'>$%.2f</td></tr>", totalAmount)
	orderDetails += "</table>"

	// Send email
	err = c.emailService.SendOrderConfirmationEmail(userEmail, userName, strconv.Itoa(orderNumber), orderDetails, totalAmount)
	if err != nil {
		fmt.Printf("Error sending order confirmation email: %v\n", err)
	}
}