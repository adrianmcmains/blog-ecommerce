package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/adrianmcmains/blog-ecommerce/internal/models"
	"github.com/adrianmcmains/blog-ecommerce/internal/service"
	"github.com/gin-gonic/gin"
)

// Common error responses
var (
	ErrInvalidInput        = errors.New("invalid input provided")
	ErrInternalServer      = errors.New("internal server error")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrUnauthorized        = errors.New("unauthorized")
)

// AuthHandler handles authentication related requests
type AuthHandler struct {
	services *service.Service
	logger   *log.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(services *service.Service, logger *log.Logger) *AuthHandler {
	return &AuthHandler{
		services: services,
		logger:   logger,
	}
}

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.RegisterInput true "User registration details"
// @Success 201 {object} map[string]interface{} "User created with auth token"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 409 {object} map[string]string "User already exists"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Printf("Invalid register input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidInput.Error()})
		return
	}

	// Validate input
	if err := validateRegisterInput(input); err != nil {
		h.logger.Printf("Register validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create user
	user, err := h.services.Auth.Register(c.Request.Context(), input)
	if err != nil {
		// Check for specific errors like duplicate email
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		
		h.logger.Printf("Register failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}

	// Generate auth token
	token, err := h.services.Auth.GenerateToken(user)
	if err != nil {
		h.logger.Printf("Token generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Return user data (excluding sensitive fields) and token
	c.JSON(http.StatusCreated, gin.H{
		"user": sanitizeUser(user),
		"token": token,
	})
}

// Login handles user login
// @Summary Log in a user
// @Description Authenticate a user and return a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.LoginInput true "User login credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Authentication failed"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Printf("Invalid login input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidInput.Error()})
		return
	}

	// Validate input
	if err := validateLoginInput(input); err != nil {
		h.logger.Printf("Login validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Authenticate and get token
	user, token, err := h.services.Auth.Login(c.Request.Context(), input)
	if err != nil {
		h.logger.Printf("Login failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrAuthenticationFailed.Error()})
		return
	}

	// Return token and user info
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": sanitizeUser(user),
	})
}

// GetCurrentUser returns the current authenticated user
// @Summary Get current user
// @Description Get the currently authenticated user's profile
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.User "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
		return
	}

	user, err := h.services.Auth.GetUserByID(c.Request.Context(), userID.(string))
	if err != nil {
		h.logger.Printf("Get current user failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}

	// Return user data (excluding sensitive fields)
	c.JSON(http.StatusOK, sanitizeUser(user))
}

// RefreshToken handles token refresh requests
// @Summary Refresh auth token
// @Description Get a new JWT token using a refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param input body models.RefreshTokenInput true "Refresh token"
// @Success 200 {object} map[string]string "New JWT token"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Invalid refresh token"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var input models.RefreshTokenInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Printf("Invalid refresh token input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidInput.Error()})
		return
	}

	// Validate and refresh the token
	newToken, err := h.services.Auth.RefreshToken(c.Request.Context(), input.RefreshToken)
	if err != nil {
		h.logger.Printf("Token refresh failed: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": newToken})
}

// Logout handles user logout
// @Summary Log out a user
// @Description Invalidate the user's current token
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get token from authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
		return
	}

	// Extract token from "Bearer <token>" format
	tokenParts := strings.Split(authHeader, " ")
	if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
		return
	}
	token := tokenParts[1]

	// Invalidate the token
	if err := h.services.Auth.InvalidateToken(c.Request.Context(), token); err != nil {
		h.logger.Printf("Logout failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// ChangePassword handles password changes
// @Summary Change user password
// @Description Change the authenticated user's password
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param input body models.ChangePasswordInput true "Password change details"
// @Success 200 {object} map[string]string "Password changed successfully"
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 401 {object} map[string]string "Unauthorized or incorrect current password"
// @Failure 500 {object} map[string]string "Server error"
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": ErrUnauthorized.Error()})
		return
	}

	var input models.ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.logger.Printf("Invalid change password input: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": ErrInvalidInput.Error()})
		return
	}

	// Validate input
	if err := validateChangePasswordInput(input); err != nil {
		h.logger.Printf("Change password validation failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Change password
	err := h.services.Auth.ChangePassword(c.Request.Context(), userID.(string), input)
	if err != nil {
		if strings.Contains(err.Error(), "incorrect password") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
			return
		}
		
		h.logger.Printf("Change password failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": ErrInternalServer.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// sanitizeUser removes sensitive information from user for safe return
// We're implementing this in the handler to avoid circular imports
func sanitizeUser(user *models.User) map[string]interface{} {
	return map[string]interface{}{
		"id":         user.ID,
		"name":       user.FirstName,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}

// Input validation functions

func validateRegisterInput(input models.RegisterInput) error {
	if input.Email == "" {
		return errors.New("email is required")
	}
	if !isValidEmail(input.Email) {
		return errors.New("invalid email format")
	}
	if input.Password == "" {
		return errors.New("password is required")
	}
	if len(input.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if input.FirstName == "" {
		return errors.New("name is required")
	}
	return nil
}

func validateLoginInput(input models.LoginInput) error {
	if input.Email == "" {
		return errors.New("email is required")
	}
	if input.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

func validateChangePasswordInput(input models.ChangePasswordInput) error {
	if input.CurrentPassword == "" {
		return errors.New("current password is required")
	}
	if input.NewPassword == "" {
		return errors.New("new password is required")
	}
	if len(input.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters")
	}
	if input.CurrentPassword == input.NewPassword {
		return errors.New("new password must be different from current password")
	}
	return nil
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	// Simple validation - in production, consider a more robust solution
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}