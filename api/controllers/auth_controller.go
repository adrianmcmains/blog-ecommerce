package controllers

import (
	"github.com/adrianmcmains/blog-ecommerce/api/middleware"
	"github.com/adrianmcmains/blog-ecommerce/api/models"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles authentication related requests
type AuthController struct {
	db *sql.DB
}

// NewAuthController creates a new auth controller
func NewAuthController(db *sql.DB) *AuthController {
	return &AuthController{db: db}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expiresAt"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents basic user information
type UserInfo struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// Login handles user login
func (c *AuthController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user by email
	var user models.User
	err := c.db.QueryRow(`
		SELECT id, name, email, password, role 
		FROM users 
		WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Calculate token expiration time
	var tokenExpHours int64 = 24
	expiresAt := time.Now().Add(time.Duration(tokenExpHours) * time.Hour)

	// Return token and user info
	ctx.JSON(http.StatusOK, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	})
}

// Register handles user registration
func (c *AuthController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var count int
	err := c.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = $1`, req.Email).Scan(&count)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	if count > 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Default role for new users
	role := "customer"

	// Insert user into database
	var userID int
	err = c.db.QueryRow(`
		INSERT INTO users (name, email, password, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, req.Name, req.Email, string(hashedPassword), role).Scan(&userID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(userID, req.Email, role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Calculate token expiration time
	var tokenExpHours int64 = 24
	expiresAt := time.Now().Add(time.Duration(tokenExpHours) * time.Hour)

	// Return token and user info
	ctx.JSON(http.StatusCreated, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: UserInfo{
			ID:    userID,
			Name:  req.Name,
			Email: req.Email,
			Role:  role,
		},
	})
}

// GetCurrentUser returns the current authenticated user
func (c *AuthController) GetCurrentUser(ctx *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Get user from database
	var user models.User
	err := c.db.QueryRow(`
		SELECT id, name, email, role 
		FROM users 
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.Role)

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Return user info
	ctx.JSON(http.StatusOK, UserInfo{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	})
}

// ChangePassword handles password change requests
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := ctx.Get("userId")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse request body
	var req struct {
		CurrentPassword string `json:"currentPassword" binding:"required"`
		NewPassword     string `json:"newPassword" binding:"required,min=8"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get current password from database
	var currentPasswordHash string
	err := c.db.QueryRow(`SELECT password FROM users WHERE id = $1`, userID).Scan(&currentPasswordHash)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(currentPasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password in database
	_, err = c.db.Exec(`
		UPDATE users 
		SET password = $1, updated_at = NOW() 
		WHERE id = $2
	`, string(hashedPassword), userID)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// Logout is a placeholder for client-side logout
// Server-side, we don't need to do anything as JWT tokens are stateless
func (c *AuthController) Logout(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}