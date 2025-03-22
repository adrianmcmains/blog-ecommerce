// File: api/controllers/payment_controller.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentController handles payment-related routes
type PaymentController struct {
	DB              *sql.DB
	EversendAPIKey  string
	EversendBaseURL string
	PayPalClientID  string
	PayPalSecret    string
	PayPalBaseURL   string
	CallbackURL     string
	WebhookSecret   string
}

// PaymentProvider represents available payment providers
type PaymentProvider string

const (
	PaymentProviderEversend PaymentProvider = "eversend"
	PaymentProviderPayPal   PaymentProvider = "paypal"
)

// NewPaymentController creates a new payment controller
func NewPaymentController(db *sql.DB) (*PaymentController, error) {
	eversendAPIKey := os.Getenv("EVERSEND_API_KEY")
	eversendBaseURL := os.Getenv("EVERSEND_BASE_URL")
	if eversendBaseURL == "" {
		eversendBaseURL = "https://api.eversend.co/v1"
	}

	paypalClientID := os.Getenv("PAYPAL_CLIENT_ID")
	paypalSecret := os.Getenv("PAYPAL_SECRET")
	paypalBaseURL := os.Getenv("PAYPAL_BASE_URL")
	if paypalBaseURL == "" {
		paypalBaseURL = "https://api-m.sandbox.paypal.com" // Sandbox by default
	}

	callbackURL := os.Getenv("PAYMENT_CALLBACK_URL")
	webhookSecret := os.Getenv("PAYMENT_WEBHOOK_SECRET")

	// Validate required credentials
	if eversendAPIKey == "" && (paypalClientID == "" || paypalSecret == "") {
		return nil, errors.New("at least one payment provider must be configured")
	}

	if callbackURL == "" {
		return nil, errors.New("payment callback URL must be set")
	}

	return &PaymentController{
		DB:              db,
		EversendAPIKey:  eversendAPIKey,
		EversendBaseURL: eversendBaseURL,
		PayPalClientID:  paypalClientID,
		PayPalSecret:    paypalSecret,
		PayPalBaseURL:   paypalBaseURL,
		CallbackURL:     callbackURL,
		WebhookSecret:   webhookSecret,
	}, nil
}

// PaymentMethodInfo contains payment method details
type PaymentMethodInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Provider    string `json:"provider"`
	Enabled     bool   `json:"enabled"`
}

// InitiatePaymentRequest contains payment initialization data
type InitiatePaymentRequest struct {
	OrderID       string `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	Currency      string `json:"currency" binding:"required"`
	ReturnURL     string `json:"return_url"`
	CancelURL     string `json:"cancel_url"`
}

// GetPaymentMethods returns available payment methods
func (c *PaymentController) GetPaymentMethods(ctx *gin.Context) {
	methods := []PaymentMethodInfo{}

	// Add Eversend payment methods if configured
	if c.EversendAPIKey != "" {
		methods = append(methods, PaymentMethodInfo{
			ID:          "eversend_card",
			Name:        "Credit/Debit Card",
			Description: "Pay with Visa, Mastercard, or other credit/debit cards",
			ImageURL:    "/images/payment/card.png",
			Provider:    string(PaymentProviderEversend),
			Enabled:     true,
		})
		methods = append(methods, PaymentMethodInfo{
			ID:          "eversend_mobile",
			Name:        "Mobile Money",
			Description: "Pay with Mobile Money",
			ImageURL:    "/images/payment/mobile.png",
			Provider:    string(PaymentProviderEversend),
			Enabled:     true,
		})
	}

	// Add PayPal payment methods if configured
	if c.PayPalClientID != "" && c.PayPalSecret != "" {
		methods = append(methods, PaymentMethodInfo{
			ID:          "paypal",
			Name:        "PayPal",
			Description: "Pay with your PayPal account",
			ImageURL:    "/images/payment/paypal.png",
			Provider:    string(PaymentProviderPayPal),
			Enabled:     true,
		})
	}

	ctx.JSON(http.StatusOK, gin.H{"payment_methods": methods})
}

// InitiatePayment starts the payment process
func (c *PaymentController) InitiatePayment(ctx *gin.Context) {
	var req InitiatePaymentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get order details
	var orderTotal float64
	var orderStatus string
	err := c.DB.QueryRow(`
		SELECT total, status FROM orders WHERE id = $1
	`, req.OrderID).Scan(&orderTotal, &orderStatus)

	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if orderStatus != "pending" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Order is not in a valid state for payment"})
		return
	}

	// Create payment record
	paymentID := uuid.New().String()
	_, err = c.DB.Exec(`
		INSERT INTO payments (
			id, order_id, amount, currency, payment_method, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
	`, paymentID, req.OrderID, orderTotal, req.Currency, req.PaymentMethod, "initiated", time.Now())

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment record"})
		return
	}

	// Determine payment provider and process accordingly
	var paymentURL string
	var providerRef string

	switch {
	case strings.HasPrefix(req.PaymentMethod, "eversend"):
		paymentURL, providerRef, err = c.processEversendPayment(paymentID, req.OrderID, orderTotal, req.Currency, req.PaymentMethod, req.ReturnURL)
	case req.PaymentMethod == "paypal":
		paymentURL, providerRef, err = c.processPayPalPayment(paymentID, req.OrderID, orderTotal, req.Currency, req.ReturnURL, req.CancelURL)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported payment method"})
		return
	}

	if err != nil {
		// Update payment status to failed
		c.DB.Exec(`
			UPDATE payments 
			SET status = 'failed', error_message = $1, updated_at = $2
			WHERE id = $3
		`, err.Error(), time.Now(), paymentID)
		
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Payment initiation failed: %v", err)})
		return
	}

	// Update payment with provider reference
	_, err = c.DB.Exec(`
		UPDATE payments
		SET provider_reference = $1, updated_at = $2
		WHERE id = $3
	`, providerRef, time.Now(), paymentID)

	if err != nil {
		// Non-critical error, just log it
		fmt.Printf("Failed to update payment with provider reference: %v\n", err)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"payment_id":   paymentID,
		"payment_url":  paymentURL,
		"provider_ref": providerRef,
	})
}

// processEversendPayment handles payment through Eversend
func (c *PaymentController) processEversendPayment(paymentID, orderID string, amount float64, currency, method, returnURL string) (string, string, error) {
	// Create payment request for Eversend API
	paymentType := "card"
	if method == "eversend_mobile" {
		paymentType = "mobile_money"
	}

	// Create request body
	requestBody := map[string]interface{}{
		"amount":      amount,
		"currency":    currency,
		"description": fmt.Sprintf("Payment for order %s", orderID),
		"payment_type": paymentType,
		"metadata": map[string]string{
			"payment_id": paymentID,
			"order_id":   orderID,
		},
		"callback_url": c.CallbackURL,
		"redirect_url": returnURL,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", err
	}

	// Make API request to Eversend
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/payments", c.EversendBaseURL), strings.NewReader(string(jsonData)))
	if err != nil {
		return "", "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.EversendAPIKey))

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return "", "", fmt.Errorf("eversend API error: %s", string(body))
	}

	// Parse response
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			ID          string `json:"id"`
			PaymentURL  string `json:"payment_url"`
			Status      string `json:"status"`
			ReferenceID string `json:"reference_id"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", "", err
	}

	if !response.Success {
		return "", "", errors.New("eversend payment creation failed")
	}

	return response.Data.PaymentURL, response.Data.ID, nil
}

// processPayPalPayment handles payment through PayPal
func (c *PaymentController) processPayPalPayment(paymentID, orderID string, amount float64, currency, returnURL, cancelURL string) (string, string, error) {
	// First, get an access token
	tokenURL := fmt.Sprintf("%s/v1/oauth2/token", c.PayPalBaseURL)
	client := &http.Client{}
	
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", "", err
	}
	
	req.SetBasicAuth(c.PayPalClientID, c.PayPalSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}
	
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", "", err
	}
	
	if tokenResponse.AccessToken == "" {
		return "", "", errors.New("failed to get PayPal access token")
	}
	
	// Create PayPal order
	orderURL := fmt.Sprintf("%s/v2/checkout/orders", c.PayPalBaseURL)
	
	// Format amount with proper precision
	amountStr := strconv.FormatFloat(amount, 'f', 2, 64)
	
	orderData := map[string]interface{}{
		"intent": "CAPTURE",
		"purchase_units": []map[string]interface{}{
			{
				"reference_id": orderID,
				"amount": map[string]interface{}{
					"currency_code": currency,
					"value":         amountStr,
				},
				"description": fmt.Sprintf("Payment for order %s", orderID),
				"custom_id":   paymentID,
			},
		},
		"application_context": map[string]interface{}{
			"return_url": returnURL,
			"cancel_url": cancelURL,
		},
	}
	
	jsonData, err := json.Marshal(orderData)
	if err != nil {
		return "", "", err
	}
	
	orderReq, err := http.NewRequest("POST", orderURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return "", "", err
	}
	
	orderReq.Header.Set("Content-Type", "application/json")
	orderReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))
	
	orderResp, err := client.Do(orderReq)
	if err != nil {
		return "", "", err
	}
	defer orderResp.Body.Close()
	
	orderBody, err := ioutil.ReadAll(orderResp.Body)
	if err != nil {
		return "", "", err
	}
	
	if orderResp.StatusCode != http.StatusCreated && orderResp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("PayPal API error: %s", string(orderBody))
	}
	
	var orderResponse struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Links  []struct {
			Href   string `json:"href"`
			Rel    string `json:"rel"`
			Method string `json:"method"`
		} `json:"links"`
	}
	
	if err := json.Unmarshal(orderBody, &orderResponse); err != nil {
		return "", "", err
	}
	
	// Find the approval URL
	var approvalURL string
	for _, link := range orderResponse.Links {
		if link.Rel == "approve" {
			approvalURL = link.Href
			break
		}
	}
	
	if approvalURL == "" {
		return "", "", errors.New("no approval URL found in PayPal response")
	}
	
	return approvalURL, orderResponse.ID, nil
}

// GetPaymentStatus checks the status of a payment
func (c *PaymentController) GetPaymentStatus(ctx *gin.Context) {
	paymentID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	var payment struct {
		ID               string    `json:"id"`
		OrderID          string    `json:"order_id"`
		Amount           float64   `json:"amount"`
		Currency         string    `json:"currency"`
		Status           string    `json:"status"`
		PaymentMethod    string    `json:"payment_method"`
		ProviderReference string   `json:"provider_reference"`
		CreatedAt        time.Time `json:"created_at"`
		UpdatedAt        time.Time `json:"updated_at"`
	}
	
	// Get payment details
	err := c.DB.QueryRow(`
		SELECT p.id, p.order_id, p.amount, p.currency, p.status, p.payment_method, p.provider_reference, p.created_at, p.updated_at
		FROM payments p
		JOIN orders o ON p.order_id = o.id
		WHERE p.id = $1 AND o.user_id = $2
	`, paymentID, userID).Scan(
		&payment.ID, &payment.OrderID, &payment.Amount, &payment.Currency,
		&payment.Status, &payment.PaymentMethod, &payment.ProviderReference,
		&payment.CreatedAt, &payment.UpdatedAt,
	)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}
	
	// If payment is in a non-final state, check with the provider for updates
	if payment.Status == "initiated" || payment.Status == "processing" {
		var newStatus string
		var err error
		
		// Check payment status with the appropriate provider
		if strings.HasPrefix(payment.PaymentMethod, "eversend") {
			newStatus, err = c.checkEversendPaymentStatus(payment.ProviderReference)
		} else if payment.PaymentMethod == "paypal" {
			newStatus, err = c.checkPayPalPaymentStatus(payment.ProviderReference)
		}
		
		if err == nil && newStatus != "" && newStatus != payment.Status {
			// Update payment status in the database
			_, err = c.DB.Exec(`
				UPDATE payments
				SET status = $1, updated_at = $2
				WHERE id = $3
			`, newStatus, time.Now(), paymentID)
			
			if err == nil {
				payment.Status = newStatus
				payment.UpdatedAt = time.Now()
				
				// If payment is completed, update order status
				if newStatus == "completed" {
					c.DB.Exec(`
						UPDATE orders
						SET status = 'processing', updated_at = $1
						WHERE id = $2 AND status = 'pending'
					`, time.Now(), payment.OrderID)
				}
			}
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{"payment": payment})
}

// checkEversendPaymentStatus checks the status of an Eversend payment
func (c *PaymentController) checkEversendPaymentStatus(paymentRef string) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/payments/%s", c.EversendBaseURL, paymentRef), nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.EversendAPIKey))
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("eversend API error: %s", string(body))
	}
	
	var response struct {
		Success bool `json:"success"`
		Data    struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}
	
	if !response.Success {
		return "", errors.New("eversend API returned unsuccessful response")
	}
	
	// Map Eversend status to our status
	switch response.Data.Status {
	case "pending":
		return "processing", nil
	case "successful":
		return "completed", nil
	case "failed":
		return "failed", nil
	default:
		return "", nil // Unknown status, don't update
	}
}

// checkPayPalPaymentStatus checks the status of a PayPal payment
func (c *PaymentController) checkPayPalPaymentStatus(orderID string) (string, error) {
	// First, get an access token
	tokenURL := fmt.Sprintf("%s/v1/oauth2/token", c.PayPalBaseURL)
	client := &http.Client{}
	
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return "", err
	}
	
	req.SetBasicAuth(c.PayPalClientID, c.PayPalSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}
	
	if tokenResponse.AccessToken == "" {
		return "", errors.New("failed to get PayPal access token")
	}
	
	// Get order details
	orderURL := fmt.Sprintf("%s/v2/checkout/orders/%s", c.PayPalBaseURL, orderID)
	
	orderReq, err := http.NewRequest("GET", orderURL, nil)
	if err != nil {
		return "", err
	}
	
	orderReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))
	
	orderResp, err := client.Do(orderReq)
	if err != nil {
		return "", err
	}
	defer orderResp.Body.Close()
	
	orderBody, err := ioutil.ReadAll(orderResp.Body)
	if err != nil {
		return "", err
	}
	
	if orderResp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("PayPal API error: %s", string(orderBody))
	}
	
	var orderResponse struct {
		Status string `json:"status"`
	}
	
	if err := json.Unmarshal(orderBody, &orderResponse); err != nil {
		return "", err
	}
	
	// Map PayPal status to our status
	switch orderResponse.Status {
	case "CREATED":
		return "initiated", nil
	case "SAVED":
		return "initiated", nil
	case "APPROVED":
		return "processing", nil
	case "VOIDED":
		return "canceled", nil
	case "COMPLETED":
		return "completed", nil
	default:
		return "", nil // Unknown status, don't update
	}
}

// CancelPayment cancels a payment
func (c *PaymentController) CancelPayment(ctx *gin.Context) {
	paymentID := ctx.Param("id")
	userID, _ := ctx.Get("userID")
	
	// Check if payment belongs to user and is in a cancellable state
	var paymentMethod, providerRef, orderID string
	var status string
	
	err := c.DB.QueryRow(`
		SELECT p.payment_method, p.provider_reference, p.status, p.order_id
		FROM payments p
		JOIN orders o ON p.order_id = o.id
		WHERE p.id = $1 AND o.user_id = $2
	`, paymentID, userID).Scan(&paymentMethod, &providerRef, &status, &orderID)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}
	
	// Check if payment can be canceled
	if status != "initiated" && status != "processing" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Payment cannot be canceled in its current state"})
		return
	}
	
	// Cancel payment with the appropriate provider
	var cancelErr error
	if strings.HasPrefix(paymentMethod, "eversend") {
		cancelErr = c.cancelEversendPayment(providerRef)
	} else if paymentMethod == "paypal" {
		cancelErr = c.cancelPayPalPayment(providerRef)
	}
	
	if cancelErr != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to cancel payment: %v", cancelErr)})
		return
	}
	
	// Update payment status
	_, err = c.DB.Exec(`
		UPDATE payments
		SET status = 'canceled', updated_at = $1
		WHERE id = $2
	`, time.Now(), paymentID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}
	
	ctx.JSON(http.StatusOK, gin.H{"message": "Payment canceled successfully"})
}

// cancelEversendPayment cancels an Eversend payment
func (c *PaymentController) cancelEversendPayment(paymentRef string) error {
	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/payments/%s/cancel", c.EversendBaseURL, paymentRef), nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.EversendAPIKey))
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("eversend API error: %s", string(body))
	}
	
	return nil
}

// cancelPayPalPayment cancels a PayPal payment
func (c *PaymentController) cancelPayPalPayment(orderID string) error {
	// First, get an access token
	tokenURL := fmt.Sprintf("%s/v1/oauth2/token", c.PayPalBaseURL)
	client := &http.Client{}
	
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return err
	}
	
	req.SetBasicAuth(c.PayPalClientID, c.PayPalSecret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return err
	}
	
	if tokenResponse.AccessToken == "" {
		return errors.New("failed to get PayPal access token")
	}
	
	// Cancel the order
	cancelURL := fmt.Sprintf("%s/v2/checkout/orders/%s/cancel", c.PayPalBaseURL, orderID)
	
	cancelReq, err := http.NewRequest("POST", cancelURL, nil)
	if err != nil {
		return err
	}
	
	cancelReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenResponse.AccessToken))
	cancelReq.Header.Set("Content-Type", "application/json")
	
	cancelResp, err := client.Do(cancelReq)
	if err != nil {
		return err
	}
	defer cancelResp.Body.Close()
	
	if cancelResp.StatusCode != http.StatusOK && cancelResp.StatusCode != http.StatusNoContent {
		cancelBody, _ := ioutil.ReadAll(cancelResp.Body)
		return fmt.Errorf("PayPal API error: %s", string(cancelBody))
	}
	
	return nil
}

// WebhookHandler handles payment webhooks from payment providers
func (c *PaymentController) WebhookHandler(ctx *gin.Context) {
	// Get webhook signature from headers
	signature := ctx.GetHeader("X-Webhook-Signature")
	
	// Read request body
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	
	// Determine the provider from headers or payload
	provider := ctx.GetHeader("X-Payment-Provider")
	if provider == "" {
		// Try to determine from the payload
		var genericPayload map[string]interface{}
		if err := json.Unmarshal(body, &genericPayload); err == nil {
			// Look for provider-specific fields
			if _, ok := genericPayload["event_type"]; ok {
				provider = "eversend"
			} else if _, ok := genericPayload["event_type"]; ok {
				provider = "paypal"
			}
		}
	}
	
	switch provider {
	case "eversend":
		c.handleEversendWebhook(ctx, body, signature)
	case "paypal":
		c.handlePayPalWebhook(ctx, body, signature)
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Unknown payment provider"})
	}
}

// handleEversendWebhook processes Eversend webhook events
func (c *PaymentController) handleEversendWebhook(ctx *gin.Context, body []byte, signature string) {
	// Verify webhook signature if available
	if c.WebhookSecret != "" && signature != "" {
		// Implementation of signature verification would go here
		// For example, using HMAC SHA256 to verify the signature
	}
	
	// Parse webhook payload
	var webhook struct {
		EventType string `json:"event_type"`
		Data struct {
			ID          string  `json:"id"`
			Status      string  `json:"status"`
			Amount      float64 `json:"amount"`
			Currency    string  `json:"currency"`
			Metadata    map[string]string `json:"metadata"`
		} `json:"data"`
	}
	
	if err := json.Unmarshal(body, &webhook); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}
	
	// Process based on event type
	if webhook.EventType != "payment.update" {
		// We only care about payment updates
		ctx.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}
	
	// Get payment ID from metadata
	paymentID, ok := webhook.Data.Metadata["payment_id"]
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing payment_id in metadata"})
		return
	}
	
	// Map Eversend status to our status
	var status string
	switch webhook.Data.Status {
	case "pending":
		status = "processing"
	case "successful":
		status = "completed"
	case "failed":
		status = "failed"
	default:
		ctx.JSON(http.StatusOK, gin.H{"status": "unknown_status"})
		return
	}
	
	// Update payment in database
	var orderID string
	err := c.DB.QueryRow(`
		SELECT order_id FROM payments
		WHERE id = $1 AND provider_reference = $2
	`, paymentID, webhook.Data.ID).Scan(&orderID)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}
	
	// Update payment status
	_, err = c.DB.Exec(`
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, status, time.Now(), paymentID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}
	
	// If payment is completed, update order status
	if status == "completed" {
		_, err = c.DB.Exec(`
			UPDATE orders
			SET status = 'processing', updated_at = $1
			WHERE id = $2
		`, time.Now(), orderID)
		
		if err != nil {
			// Non-critical error, just log it
			fmt.Printf("Failed to update order status: %v\n", err)
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{"status": "processed"})
}

// handlePayPalWebhook processes PayPal webhook events
func (c *PaymentController) handlePayPalWebhook(ctx *gin.Context, body []byte, signature string) {
	// Verify webhook signature if available
	if c.WebhookSecret != "" && signature != "" {
		// Implementation of signature verification would go here
	}
	
	// Parse webhook payload
	var webhook struct {
		EventType string `json:"event_type"`
		Resource struct {
			ID         string `json:"id"`
			Status     string `json:"status"`
			CustomID   string `json:"custom_id"` // This contains our payment ID
			PurchaseUnits []struct {
				ReferenceID string `json:"reference_id"` // This contains our order ID
			} `json:"purchase_units"`
		} `json:"resource"`
	}
	
	if err := json.Unmarshal(body, &webhook); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
		return
	}
	
	// We only care about certain event types
	validEvents := map[string]bool{
		"PAYMENT.AUTHORIZATION.CREATED": true,
		"PAYMENT.CAPTURE.COMPLETED":     true,
		"PAYMENT.CAPTURE.DENIED":        true,
		"CHECKOUT.ORDER.APPROVED":       true,
		"CHECKOUT.ORDER.COMPLETED":      true,
	}
	
	if !validEvents[webhook.EventType] {
		ctx.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}
	
	// Get payment ID from custom_id
	paymentID := webhook.Resource.CustomID
	if paymentID == "" && len(webhook.Resource.PurchaseUnits) > 0 {
		// Try to find payment by order reference
		orderID := webhook.Resource.PurchaseUnits[0].ReferenceID
		if orderID != "" {
			err := c.DB.QueryRow(`
				SELECT id FROM payments
				WHERE order_id = $1 AND provider_reference = $2
			`, orderID, webhook.Resource.ID).Scan(&paymentID)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
				return
			}
		} else {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing payment identifier"})
			return
		}
	}
	
	// Map PayPal status to our status
	var status string
	switch webhook.Resource.Status {
	case "CREATED":
		status = "initiated"
	case "SAVED":
		status = "initiated"
	case "APPROVED":
		status = "processing"
	case "VOIDED":
		status = "canceled"
	case "COMPLETED":
		status = "completed"
	default:
		ctx.JSON(http.StatusOK, gin.H{"status": "unknown_status"})
		return
	}
	
	// Update payment in database
	var orderID string
	err := c.DB.QueryRow(`
		SELECT order_id FROM payments
		WHERE id = $1
	`, paymentID).Scan(&orderID)
	
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}
	
	// Update payment status
	_, err = c.DB.Exec(`
		UPDATE payments
		SET status = $1, updated_at = $2
		WHERE id = $3
	`, status, time.Now(), paymentID)
	
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}
	
	// If payment is completed, update order status
	if status == "completed" {
		_, err = c.DB.Exec(`
			UPDATE orders
			SET status = 'processing', updated_at = $1
			WHERE id = $2
		`, time.Now(), orderID)
		
		if err != nil {
			// Non-critical error, just log it
			fmt.Printf("Failed to update order status: %v\n", err)
		}
	}
	
	ctx.JSON(http.StatusOK, gin.H{"status": "processed"})
}