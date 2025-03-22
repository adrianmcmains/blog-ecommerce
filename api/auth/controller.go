package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Controller handles authentication related routes
type Controller struct {
	DB *gorm.DB
}

// NewController creates a new auth controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{DB: db}
}

// Login authenticates a user and returns a JWT token
func (c *Controller) Login(ctx *gin.Context) {
	var loginRequest LoginRequest
	if err := ctx.ShouldBindJSON(&loginRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := c.DB.Where("email = ?", loginRequest.Email).First(&user).Error; err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !CheckPasswordHash(loginRequest.Password, user.Password) {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := GenerateToken(user.ID.String(), user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// Register creates a new user
func (c *Controller) Register(ctx *gin.Context) {
	var registerRequest RegisterRequest
	if err := ctx.ShouldBindJSON(&registerRequest); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser User
	if err := c.DB.Where("email = ?", registerRequest.Email).First(&existingUser).Error; err == nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	hashedPassword, err := HashPassword(registerRequest.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process registration"})
		return
	}

	user := User{
		ID:        uuid.New(),
		Email:     registerRequest.Email,
		Password:  hashedPassword,
		FirstName: registerRequest.FirstName,
		LastName:  registerRequest.LastName,
		Role:      "customer",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if result := c.DB.Create(&user); result.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token, err := GenerateToken(user.ID.String(), user.Role)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"token": token,
		"user":  user.ToResponse(),
	})
}

// GetMe retrieves the current user's information
func (c *Controller) GetMe(ctx *gin.Context) {
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var user User
	if err := c.DB.First(&user, "id = ?", userID).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"user": user.ToResponse()})
}
