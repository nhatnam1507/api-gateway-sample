package usecase

import (
	"context"

	"api-gateway-sample/internal/domain/service"
	"api-gateway-sample/pkg/logger"
)

// AuthUseCase implements the use case for authentication
type AuthUseCase struct {
	authService service.AuthService
	logger      logger.Logger
}

// NewAuthUseCase creates a new AuthUseCase instance
func NewAuthUseCase(authService service.AuthService, logger logger.Logger) *AuthUseCase {
	return &AuthUseCase{
		authService: authService,
		logger:      logger,
	}
}

// GenerateToken generates an authentication token
func (uc *AuthUseCase) GenerateToken(ctx context.Context, userID string, claims map[string]interface{}) (string, error) {
	return uc.authService.GenerateToken(ctx, userID, claims)
}

// ValidateToken validates an authentication token
func (uc *AuthUseCase) ValidateToken(ctx context.Context, token string) (map[string]interface{}, error) {
	return uc.authService.ValidateToken(ctx, token)
}
