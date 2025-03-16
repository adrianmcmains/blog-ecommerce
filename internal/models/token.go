package models

import (
	"time"

	"github.com/google/uuid"
)

// Token represents a JWT token with refresh capabilities
type Token struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	UserID       string    `json:"user_id" bson:"user_id"`
	AccessToken  string    `json:"access_token,omitempty" bson:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty" bson:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" bson:"expires_at"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	IsRevoked    bool      `json:"is_revoked" bson:"is_revoked"`
}

// NewToken creates a new token with default values
func NewToken(userID, accessToken, refreshToken string, expiresIn time.Duration) *Token {
	now := time.Now()
	return &Token{
		ID:           uuid.New().String(),
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(expiresIn),
		CreatedAt:    now,
		IsRevoked:    false,
	}
}