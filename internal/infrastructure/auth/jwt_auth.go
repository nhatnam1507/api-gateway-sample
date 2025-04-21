package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"api-gateway-sample/internal/domain/entity"
	"api-gateway-sample/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth implements the AuthService interface using JWT
type JWTAuth struct {
	secretKey  []byte
	issuer     string
	expiration time.Duration
	logger     logger.Logger
}

// NewJWTAuth creates a new JWTAuth instance
func NewJWTAuth(secretKey []byte, issuer string, expiration time.Duration, logger logger.Logger) *JWTAuth {
	return &JWTAuth{
		secretKey:  secretKey,
		issuer:     issuer,
		expiration: expiration,
		logger:     logger,
	}
}

// getAuthToken extracts the token from the Authorization header
func getAuthToken(headers map[string][]string) string {
	if authHeaders, ok := headers["Authorization"]; ok && len(authHeaders) > 0 {
		authHeader := authHeaders[0]
		if strings.HasPrefix(authHeader, "Bearer ") {
			return strings.TrimPrefix(authHeader, "Bearer ")
		}
		return authHeader
	}
	return ""
}

// Authenticate authenticates a request
func (a *JWTAuth) Authenticate(ctx context.Context, request *entity.Request) (bool, string, error) {
	tokenString := getAuthToken(request.Headers)
	if tokenString == "" {
		return false, "", nil
	}

	claims, err := a.ValidateToken(ctx, tokenString)
	if err != nil {
		return false, "", err
	}

	userID, ok := claims["sub"].(string)
	if !ok {
		return false, "", fmt.Errorf("invalid user ID in token")
	}

	return true, userID, nil
}

// Authorize authorizes a request for a specific service and endpoint
func (a *JWTAuth) Authorize(ctx context.Context, request *entity.Request, service *entity.Service, endpoint *entity.Endpoint) error {
	if !endpoint.AuthRequired {
		return nil
	}

	tokenString := getAuthToken(request.Headers)
	if tokenString == "" {
		return fmt.Errorf("authorization required")
	}

	claims, err := a.ValidateToken(ctx, tokenString)
	if err != nil {
		return err
	}

	// Check roles/permissions from claims
	roles, ok := claims["roles"].([]interface{})
	if !ok {
		return fmt.Errorf("invalid roles in token")
	}

	// Simple role-based authorization
	hasAccess := false
	for _, role := range roles {
		if roleStr, ok := role.(string); ok {
			if roleStr == "admin" || roleStr == service.Name+":"+endpoint.Path {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return fmt.Errorf("unauthorized: insufficient permissions")
	}

	return nil
}

// GenerateToken generates an authentication token
func (a *JWTAuth) GenerateToken(ctx context.Context, userID string, claims map[string]interface{}) (string, error) {
	now := time.Now()
	tokenClaims := jwt.MapClaims{
		"iss": a.issuer,
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(a.expiration).Unix(),
	}

	// Add custom claims
	for k, v := range claims {
		tokenClaims[k] = v
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)
	return token.SignedString(a.secretKey)
}

// ValidateToken validates an authentication token
func (a *JWTAuth) ValidateToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}
