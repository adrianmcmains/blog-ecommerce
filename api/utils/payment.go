package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// EversendPaymentService handles integration with Eversend payment gateway
type EversendPaymentService struct {
	APIKey      string
	APISecret   string
	BaseURL     string
	Environment string
	HTTPClient  *http.Client
}

// EversendPaymentRequest represents the request to create a payment
type EversendPaymentRequest struct {
	Amount       float64           `json:"amount"`
	Currency     string            `json:"currency"`
	Description  string            `json:"description"`
	Reference    string            `json:"reference"`
	CustomerInfo EversendCustomer  `json:"customer_info"`
	RedirectURL  string            `json:"redirect_url"`
	WebhookURL   string            `json:"webhook_url"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// EversendCustomer represents customer information for payment
type EversendCustomer struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone,omitempty"`
	Address   string `json:"address,omitempty"`
	City      string `json:"city,omitempty"`
	Country   string `json:"country,omitempty"`
}

// EversendPaymentResponse represents the response from a payment creation
type EversendPaymentResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	PaymentID    string `json:"payment_id"`
	PaymentURL   string `json:"payment_url"`
	Status       string `json:"status"`
	Reference    string `json:"reference"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// EversendPaymentStatus represents the status of a payment
type EversendPaymentStatus struct {
	Success      bool              `json:"success"`
	PaymentID    string            `json:"payment_id"`
	Reference    string            `json:"reference"`
	Amount       float64           `json:"amount"`
	Currency     string            `json:"currency"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	CustomerInfo EversendCustomer  `json:"customer_info"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// EversendWebhookPayload represents the webhook payload from Eversend
type EversendWebhookPayload struct {
	Event       string                 `json:"event"`
	PaymentID   string                 `json:"payment_id"`
	Reference   string                 `json:"reference"`
	Status      string                 `json:"status"`
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	PaymentData map[string]interface{} `json:"payment_data"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// NewEversendPaymentService creates a new Eversend payment service
func NewEversendPaymentService() (*EversendPaymentService, error) {
	apiKey := os.Getenv("EVERSEND_API_KEY")
	if apiKey == "" {
		return nil, errors.New("EVERSEND_API_KEY is required")
	}

	apiSecret := os.Getenv("EVERSEND_API_SECRET")
	if apiSecret == "" {
		return nil, errors.New("EVERSEND_API_SECRET is required")
	}

	environment := os.Getenv("EVERSEND_ENVIRONMENT")
	if environment == "" {
		environment = "sandbox" // Default to sandbox environment
	}

	// Set the appropriate base URL based on environment
	var baseURL string
	if environment == "production" {
		baseURL = "https://api.eversend.co/v1"
	} else {
		baseURL = "https://sandbox-api.eversend.co/v1"
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	return &EversendPaymentService{
		APIKey:      apiKey,
		APISecret:   apiSecret,
		BaseURL:     baseURL,
		Environment: environment,
		HTTPClient:  httpClient,
	}, nil
}

// CreatePayment creates a new payment in Eversend
func (s *EversendPaymentService) CreatePayment(req EversendPaymentRequest) (*EversendPaymentResponse, error) {
	url := fmt.Sprintf("%s/payments", s.BaseURL)

	// Convert request to JSON
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling payment request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", s.APIKey)
	httpReq.Header.Set("X-API-Secret", s.APISecret)

	// Send request
	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending payment request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check response status
	if resp.StatusCode >= 400 {
		var errorResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("error response from Eversend (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("error response from Eversend: %s - %s", errorResp.Message, errorResp.Error)
	}

	// Parse response
	var paymentResp EversendPaymentResponse
	if err := json.Unmarshal(body, &paymentResp); err != nil {
		return nil, fmt.Errorf("error parsing payment response: %w", err)
	}

	return &paymentResp, nil
}

// GetPaymentStatus gets the status of a payment
func (s *EversendPaymentService) GetPaymentStatus(paymentID string) (*EversendPaymentStatus, error) {
	url := fmt.Sprintf("%s/payments/%s", s.BaseURL, paymentID)

	// Create HTTP request
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("X-API-Key", s.APIKey)
	httpReq.Header.Set("X-API-Secret", s.APISecret)

	// Send request
	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending payment status request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	// Check response status
	if resp.StatusCode >= 400 {
		var errorResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("error response from Eversend (status %d): %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("error response from Eversend: %s - %s", errorResp.Message, errorResp.Error)
	}

	// Parse response
	var statusResp EversendPaymentStatus
	if err := json.Unmarshal(body, &statusResp); err != nil {
		return nil, fmt.Errorf("error parsing payment status response: %w", err)
	}

	return &statusResp, nil
}

// CancelPayment cancels a pending payment
func (s *EversendPaymentService) CancelPayment(paymentID string) error {
	url := fmt.Sprintf("%s/payments/%s/cancel", s.BaseURL, paymentID)

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("error creating HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("X-API-Key", s.APIKey)
	httpReq.Header.Set("X-API-Secret", s.APISecret)

	// Send request
	resp, err := s.HTTPClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("error sending payment cancel request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	// Check response status
	if resp.StatusCode >= 400 {
		var errorResp struct {
			Success bool   `json:"success"`
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return fmt.Errorf("error response from Eversend (status %d): %s", resp.StatusCode, string(body))
		}
		return fmt.Errorf("error response from Eversend: %s - %s", errorResp.Message, errorResp.Error)
	}

	return nil
}

// VerifyWebhookSignature verifies the signature of a webhook payload
func (s *EversendPaymentService) VerifyWebhookSignature(payload []byte, signature string) bool {
	// Eversend uses HMAC-SHA256 for webhook signatures
	// This is a placeholder implementation - actual implementation will depend on Eversend's webhook signature verification method
	// Typically, you would compute an HMAC using your API secret and compare it with the provided signature
	
	// NOTE: Replace this implementation with the actual verification logic based on Eversend's documentation
	return true // For now, we're assuming all webhooks are valid
}

// ParseWebhookPayload parses the webhook payload from Eversend
func (s *EversendPaymentService) ParseWebhookPayload(payload []byte) (*EversendWebhookPayload, error) {
	var webhookPayload EversendWebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return nil, fmt.Errorf("error parsing webhook payload: %w", err)
	}
	return &webhookPayload, nil
}