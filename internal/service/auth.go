package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/adrianmcmains/blog-ecommerce/internal/models"
	"github.com/adrianmcmains/blog-ecommerce/internal/repository"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Common errors
var (
	ErrUserExists      = errors.New("user with this email already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("incorrect password")
	ErrInvalidToken    = errors.New("invalid or expired token")
)

// AuthService handles authentication related business logic
type AuthService struct {
	userRepo   repository.UserRepository
	tokenRepo  repository.TokenRepository
	jwtSecret  string
	jwtTTL     time.Duration
	refreshTTL time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	jwtSecret string,
	jwtTTL time.Duration,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtSecret:  jwtSecret,
		jwtTTL:     jwtTTL,
		refreshTTL: refreshTTL,
	}
}

// Register registers a new user
func (s *AuthService) Register(ctx context.Context, input models.RegisterInput) (*models.User, error) {
	// Check if user already exists
	existing, _ := s.userRepo.GetByEmail(ctx, input.Email)
	if existing != nil {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		FirstName:    input.FirstName,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
		Role:         "user", // Default role
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Login handles user login
func (s *AuthService) Login(ctx context.Context, input models.LoginInput) (*models.User, string, error) {
	// Find user by email
	user, err := s.userRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, "", ErrUserNotFound
	}

	// Verify password
	if !user.CheckPassword(input.Password) {
		return nil, "", ErrInvalidPassword
	}

	// Generate tokens
	accessToken, refreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Store token in repository
	tokenModel := &models.Token{
		UserID:       user.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
	}
	
	err = s.tokenRepo.Create(ctx, tokenModel)
	if err != nil {
		return nil, "", fmt.Errorf("failed to store token: %w", err)
	}

	return user, accessToken, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GenerateToken generates a JWT token for a user
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	token, _, err := s.generateTokenPair(user)
	return token, err
}

// RefreshToken refreshes an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	// Verify the refresh token exists and is valid
	token, err := s.tokenRepo.GetByRefreshToken(ctx, refreshToken)
	if err != nil || token.IsRevoked || token.ExpiresAt.Before(time.Now()) {
		return "", ErrInvalidToken
	}

	// Get the associated user
	user, err := s.userRepo.GetByID(ctx, token.UserID)
	if err != nil {
		return "", ErrUserNotFound
	}

	// Generate new access token
	newAccessToken, newRefreshToken, err := s.generateTokenPair(user)
	if err != nil {
		return "", fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Revoke the old token
	err = s.tokenRepo.Revoke(ctx, token.ID)
	if err != nil {
		return "", fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Store new token
	newToken := &models.Token{
		UserID:       user.ID.String(),
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    time.Now().Add(s.refreshTTL),
	}
	
	err = s.tokenRepo.Create(ctx, newToken)
	if err != nil {
		return "", fmt.Errorf("failed to store new token: %w", err)
	}

	return newAccessToken, nil
}

// ValidateToken validates a token and returns it if valid
func (s *AuthService) ValidateToken(ctx context.Context, accessToken string) (*models.Token, error) {
	// First parse the token to check its validity
	_, err := s.parseToken(accessToken)
	if err != nil {
		return nil, err
	}

	// If the token is valid, check if it exists in the database and is not revoked
	token, err := s.tokenRepo.GetByAccessToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}

	if token.IsRevoked {
		return nil, ErrInvalidToken
	}

	return token, nil
}

// InvalidateToken marks a token as revoked
func (s *AuthService) InvalidateToken(ctx context.Context, accessToken string) error {
	// Extract claims from token to get user ID
	claims, err := s.extractTokenClaims(accessToken)
	if err != nil {
		return ErrInvalidToken
	}

	// Find the token by access token
	token, err := s.tokenRepo.GetByAccessToken(ctx, accessToken)
	if err != nil {
		return ErrInvalidToken
	}

	// Verify token belongs to user
	userID, ok := claims["user_id"].(string)
	if !ok || userID != token.UserID {
		return ErrInvalidToken
	}

	// Revoke the token
	return s.tokenRepo.Revoke(ctx, token.ID)
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, userID string, input models.ChangePasswordInput) error {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify current password
	if !user.CheckPassword(input.CurrentPassword) {
		return ErrInvalidPassword
	}

	// Update password
	err = user.SetPassword(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	
	err = s.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Revoke all tokens for this user (optional but recommended)
	err = s.tokenRepo.RevokeAllForUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}

	return nil
}

// Private helper methods

// generateTokenPair creates a new JWT access token and refresh token
func (s *AuthService) generateTokenPair(user *models.User) (string, string, error) {
	// Create access token
	accessClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(s.jwtTTL).Unix(),
		"jti":     uuid.New().String(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	// Create refresh token
	refreshClaims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"exp":     time.Now().Add(s.refreshTTL).Unix(),
		"jti":     uuid.New().String(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}

// parseToken parses and validates a JWT token
func (s *AuthService) parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})
}

// extractTokenClaims extracts claims from a JWT token
func (s *AuthService) extractTokenClaims(tokenString string) (jwt.MapClaims, error) {
	token, err := s.parseToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}