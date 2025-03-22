package middleware

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// User roles
const (
	RoleAdmin       = "admin"
	RoleCustomer    = "customer"
	RoleContributor = "contributor"
)

// JWTClaims defines the claims in JWT token
type JWTClaims struct {
	UserID int    `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware returns middleware for JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		token, err := extractToken(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
			c.Abort()
			return
		}

		// Validate token
		claims, err := validateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: " + err.Error()})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set("userId", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

// RoleMiddleware returns middleware for checking user roles
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (set by AuthMiddleware)
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user role not found"})
			c.Abort()
			return
		}

		// Check if the user's role is in the list of allowed roles
		userRole := role.(string)
		for _, allowedRole := range allowedRoles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		// If the user's role is not allowed, return forbidden error
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
		c.Abort()
	}
}

// RequireRole returns a middleware that checks if the user has any of the specified roles
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user role from context (this should be set by AuthMiddleware)
		role, exists := c.Get("userRole")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: user role not found"})
			c.Abort()
			return
		}

		// Check if the user's role is in the list of allowed roles
		userRole := role.(string)
		for _, allowedRole := range roles {
			if userRole == allowedRole {
				c.Next()
				return
			}
		}

		// If the user's role is not allowed, return forbidden error
		c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
		c.Abort()
	}
}

// ExtractToken extracts JWT token from request header
func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Check if Authorization header has the right format
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("authorization header format must be Bearer {token}")
	}

	return parts[1], nil
}

// ValidateToken validates JWT token and returns the claims
func validateToken(tokenString string) (*JWTClaims, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET is not set")
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		},
	)
	if err != nil {
		return nil, err
	}

	// Type assert claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is expired
	if time.Now().Unix() > claims.ExpiresAt.Unix() {
		return nil, errors.New("token expired")
	}

	return claims, nil
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(userID int, email, role string) (string, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET is not set")
	}

	// Get token expiration time from environment or use default (24 hours)
	var tokenExpHours int64 = 24
	if expStr := os.Getenv("JWT_EXPIRATION_HOURS"); expStr != "" {
		if exp, err := time.ParseDuration(expStr + "h"); err == nil {
			tokenExpHours = int64(exp.Hours())
		}
	}

	// Set expiration time
	expirationTime := time.Now().Add(time.Duration(tokenExpHours) * time.Hour)

	// Create claims
	claims := &JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// CSRFMiddleware adds CSRF protection to routes
func CSRFMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check POST, PUT, DELETE, PATCH requests
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Get CSRF token from header
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "CSRF token is required"})
			c.Abort()
			return
		}

		// Validate CSRF token
		// In a real implementation, you would compare the token with one stored in the user's session
		// For simplicity, we'll just check that it's not empty
		if csrfToken == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy (CSP)
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' https://cdnjs.cloudflare.com; style-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; img-src 'self' data:; font-src 'self' https://cdnjs.cloudflare.com; connect-src 'self'")
		
		// HTTP Strict Transport Security (HSTS)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		
		// X-Content-Type-Options
		c.Header("X-Content-Type-Options", "nosniff")
		
		// X-Frame-Options
		c.Header("X-Frame-Options", "DENY")
		
		// X-XSS-Protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Referrer-Policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// Feature-Policy
		c.Header("Feature-Policy", "camera 'none'; microphone 'none'; geolocation 'none'")
		
		c.Next()
	}
}

// RateLimiter implements a simple rate limiting middleware
func RateLimiter(requests int, duration time.Duration) gin.HandlerFunc {
	type client struct {
		count    int
		lastSeen time.Time
	}
	
	clients := make(map[string]*client)
	mu := &sync.Mutex{}
	
	go func() {
		// Clean up old entries periodically
		for {
			time.Sleep(duration)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > duration {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	
	return func(c *gin.Context) {
		ip := c.ClientIP()
		
		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{count: 0, lastSeen: time.Now()}
		}
		
		if clients[ip].count >= requests {
			if time.Since(clients[ip].lastSeen) < duration {
				mu.Unlock()
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
				c.Abort()
				return
			}
			// Reset counter after duration
			clients[ip].count = 0
		}
		
		clients[ip].count++
		clients[ip].lastSeen = time.Now()
		mu.Unlock()
		
		c.Next()
	}
}