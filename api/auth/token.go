// File: api/auth/token.go
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Secret key for JWT signing. In production, use environment variables.
var secretKey = []byte("your-secret-key-replace-in-production")

// CustomClaims represents the JWT claims
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken creates a new JWT token for the user
func GenerateToken(userID string, role string) (string, error) {
	// Create token with claims
	claims := CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

// ValidateToken validates the JWT token and returns the claims
func ValidateToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the alg is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}